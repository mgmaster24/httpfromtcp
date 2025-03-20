package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mgmaster24/httpfromtcp/internal/headers"
)

type Request struct {
	Headers     headers.Headers
	RequestLine RequestLine
	state       parserStatw
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	crlf       = "\r\n"
	bufferSize = 8
)

type parserStatw int

const (
	Initialized    parserStatw = 0
	ParsingHeaders parserStatw = 1
	Done           parserStatw = 42
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state:   Initialized,
		Headers: headers.NewHeaders(),
	}

	readToIndex := 0
	buf := make([]byte, bufferSize, bufferSize)
	for request.state != Done {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(io.EOF, err) {
				if request.state != Done {
					return nil, fmt.Errorf(
						"incomplete request, in state: %d, read n bytes on EOF: %d",
						request.state,
						bytesRead,
					)
				}
				break
			}
			return nil, err
		}

		readToIndex += bytesRead
		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[bytesParsed:readToIndex])
		readToIndex -= bytesParsed
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case Initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = ParsingHeaders
		return n, nil
	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = Done
		}
		return n, nil
	case Done:
		return 0, fmt.Errorf("trying to read data in Done state")
	default:
		return 0, fmt.Errorf("unknown status %v", r.state)
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + len(crlf), nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request line: %s", str)
	}

	method := parts[0]
	if method != strings.ToUpper(method) {
		return nil, fmt.Errorf("invalid method: %s", str)
	}

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", str)
	}

	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", str)
	}

	return &RequestLine{
		HttpVersion:   version,
		Method:        method,
		RequestTarget: parts[1],
	}, nil
}
