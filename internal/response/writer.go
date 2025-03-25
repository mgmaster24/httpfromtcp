package response

import (
	"fmt"
	"io"

	"github.com/mgmaster24/httpfromtcp/internal/headers"
)

type writerState int

const (
	StatusLine writerState = 0
	Headers    writerState = 1
	Body       writerState = 2
	Done       writerState = 42
)

type Writer struct {
	state  writerState
	Writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state:  StatusLine,
		Writer: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StatusLine {
		return fmt.Errorf("writer in incorrect state, state %d", w.state)
	}

	defer func() { w.state = Headers }()
	return WriteStatusLine(w.Writer, statusCode)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != Headers {
		return fmt.Errorf("writer in incorrect state, state %d", w.state)
	}

	defer func() { w.state = Body }()
	return WriteHeaders(w.Writer, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != Body {
		return 0, fmt.Errorf("writer in incorrect state, state %d", w.state)
	}

	defer func() { w.state = Done }()
	return w.Writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	var buf []byte
	buf = fmt.Appendf(buf, "%x\r\n", len(p))
	chunkStartLen, err := w.Writer.Write(buf)
	if err != nil {
		return 0, err
	}
	chunkBodyLen, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}

	chunkEndLen, err := w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	return chunkStartLen + chunkBodyLen + chunkEndLen, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	var buf []byte
	buf = fmt.Appendf(buf, "%x\r\n", 0)
	chunkStartLen, err := w.Writer.Write(buf)
	if err != nil {
		return 0, err
	}
	chunkBodyLen, err := w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	return chunkStartLen + chunkBodyLen, nil
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
