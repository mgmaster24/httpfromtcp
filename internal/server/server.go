package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/mgmaster24/httpfromtcp/internal/request"
	"github.com/mgmaster24/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

func WriteResponse(w io.Writer, statusCode response.StatusCode, contentLen int) {
	err := response.WriteStatusLine(w, statusCode)
	if err != nil {
		log.Print("Error writing the status line")
	}

	err = response.WriteHeaders(w, response.GetDefaultHeaders(contentLen))
	if err != nil {
		log.Print("Error writing headers")
	}
}

type Handler func(w io.Writer, req *request.Request)

func Serve(port int32, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error accepting new connection %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	request, err := request.RequestFromReader(conn)
	if err != nil {
		WriteResponse(conn, response.InternalServerError, 0)
		return
	}

	var buf bytes.Buffer
	s.handler(&buf, request)

	_, err = buf.WriteTo(conn)
	if err != nil {
		log.Printf("error writing the content to the connection. err: %e", err)
	}
}
