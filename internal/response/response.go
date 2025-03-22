package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/mgmaster24/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	Unrecognized        StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusCodeResponses := make(map[StatusCode]string)
	templateString := "HTTP/1.1 %d %s\r\n"
	statusCodeResponses[Ok] = fmt.Sprintf(templateString, Ok, "OK")
	statusCodeResponses[Unrecognized] = fmt.Sprintf(templateString, Unrecognized, "Bad Request")
	statusCodeResponses[InternalServerError] = fmt.Sprintf(
		templateString,
		InternalServerError,
		"Internal Server Error",
	)

	if resp, ok := statusCodeResponses[statusCode]; ok {
		_, err := w.Write([]byte(resp))
		return err
	}

	_, err := fmt.Fprintf(w, templateString, statusCode, "")
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))
	return err
}
