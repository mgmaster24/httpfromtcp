package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mgmaster24/httpfromtcp/internal/headers"
	"github.com/mgmaster24/httpfromtcp/internal/request"
	"github.com/mgmaster24/httpfromtcp/internal/response"
	"github.com/mgmaster24/httpfromtcp/internal/server"
)

const port = 42069

const HTML400 = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const HTML500 = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const HTMLOK = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) {
	writer := response.NewWriter(w)
	hdrs := headers.NewHeaders()
	hdrs.Set("Content-Type", "text/html")

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
		url := fmt.Sprintf("https://httpbin.org%s", target)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error getting data from %s", url)
			handle(writer, &hdrs, response.InternalServerError, HTML500)
			return
		}
		defer resp.Body.Close()
		writer.WriteStatusLine(response.StatusCode(resp.StatusCode))
		hdrs.Set("Transfer-Encoding", "chunked")
		for k, v := range resp.Header {
			if k != "Content-Length" {
				hdrs.Set(k, v[0])
			}
		}

		writer.WriteHeaders(hdrs)
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				if errors.Is(io.EOF, err) {
					n, err = writer.WriteChunkedBodyDone()
					return
				}
				log.Printf("Error reading from response. err: %e", err)
				break
			}

			log.Printf("Bytes read. %d", n)
			writer.WriteChunkedBody(buf[:n])
		}
		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		handle(writer, &hdrs, response.BadRequest, HTML400)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		handle(writer, &hdrs, response.InternalServerError, HTML500)
		return
	}

	handle(writer, &hdrs, response.Ok, HTMLOK)
}

func handle(
	writer *response.Writer,
	hdrs *headers.Headers,
	statusCode response.StatusCode,
	body string,
) {
	writer.WriteStatusLine(statusCode)
	hdrs.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	writer.WriteHeaders(*hdrs)
	writer.WriteBody([]byte(body))
}
