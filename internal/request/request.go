package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	Status      Status
	RequestLine RequestLine
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

type Status int

const (
	Initialized Status = 0
	Done        Status = 1
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		Status: Initialized,
	}

	readToIndex := 0
	buf := make([]byte, bufferSize)
	for request.Status != Done {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				request.Status = Done
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
	if r.Status == Done {
		return 0, fmt.Errorf("attempting to read data in done state")
	}

	if r.Status != Initialized {
		return 0, fmt.Errorf("unknown status %v", r.Status)
	}

	requestLine, numBytes, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}

	if numBytes == 0 {
		return 0, nil
	}

	r.RequestLine = *requestLine
	r.Status = Done

	return numBytes, nil
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

	return requestLine, idx + 2, nil
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
