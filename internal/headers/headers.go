package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	headers := make(map[string]string)
	return headers
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	crlfIdx := bytes.Index(data, []byte(crlf))
	if crlfIdx == -1 {
		return 0, false, nil
	}

	if crlfIdx == 0 {
		return len(crlf), true, nil
	}

	parts := bytes.SplitN(data[:crlfIdx], []byte(":"), 2)
	fieldName := strings.ToLower(string(parts[0]))
	if fieldName != strings.TrimRight(fieldName, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", fieldName)
	}

	fieldName = strings.TrimSpace(fieldName)
	if !isValidString(fieldName) {
		return 0, false, fmt.Errorf("invalid header token found: %s", fieldName)
	}

	fieldValue := bytes.TrimSpace(parts[1])
	h.Set(fieldName, string(fieldValue))
	return crlfIdx + len(crlf), false, nil
}

func (h Headers) Set(key string, value string) {
	if val, ok := h[key]; ok {
		value = strings.Join([]string{val, value}, ", ")
	}

	h[key] = value
}

func isValidString(s string) bool {
	specialChars := "!#$%&'*+-./^_`|~"
	for _, char := range s {
		if !(unicode.IsLetter(char) || unicode.IsDigit(char) || contains(specialChars, char)) {
			return false
		}
	}
	return true
}

func contains(chars string, char rune) bool {
	for _, c := range chars {
		if c == char {
			return true
		}
	}
	return false
}
