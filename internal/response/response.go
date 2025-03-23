package response

import (
	"fmt"
	"io"

	"github.com/mgmaster24/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusCodeResponses := make(map[StatusCode]string)
	templateString := "HTTP/1.1 %d %s\r\n"
	statusCodeResponses[Ok] = fmt.Sprintf(templateString, Ok, "OK")
	statusCodeResponses[BadRequest] = fmt.Sprintf(templateString, BadRequest, "Bad Request")
	statusCodeResponses[InternalServerError] = fmt.Sprintf(
		templateString,
		InternalServerError,
		"Internal Server Error",
	)

	if resp, ok := statusCodeResponses[statusCode]; ok {
		_, err := w.Write([]byte(resp))
		return err
	}

	_, err := fmt.Fprintf(w, templateString, statusCode, "Unknown status code")
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}
