package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	headers := make(map[string]string)
	return headers
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headers := string(data)
	if !strings.Contains(headers, crlf) {
		return 0, false, nil
	}

	headers = headers[:strings.Index(headers, crlf)-1]
	if len(headers) == 0 {
		return 0, true, nil
	}

	index := strings.Index(headers, ":")
	if headers[index-1] == ' ' {
		return 0, false, fmt.Errorf("field line is in incorrect format: %s", headers)
	}

	fieldName := headers[:index]
	fieldName = strings.TrimSpace(fieldName)
	fieldValue := headers[index+1:]
	fieldValue = strings.TrimSpace(fieldValue)
	h[fieldName] = fieldValue
	return len(data) - 2, false, nil
}
