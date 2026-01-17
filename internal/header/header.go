package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

var crlf = []byte("\r\n")

var ERROR_BAD_HEADER = fmt.Errorf("Invalid Header..")

// getter and setter name and value field
func (h *Headers) Get(name string) string {

	return h.headers[strings.ToLower(name)]

}

func (h *Headers) Replace(name, value string) {

	name = strings.ToLower(name)

	h.headers[name] = value

}

func (h *Headers) Set(name, value string) {

	name = strings.ToLower(name)

	if v, ok := h.headers[name]; ok {

		h.headers[name] = fmt.Sprintf("%s, %s", v, value)
	} else {

		h.headers[name] = value
	}

}

// This function iterates over all the headers in the Headers map and calls a callback function for each header name-value pair.
func (h *Headers) ForEach(cb func(n, v string)) {

	for n, v := range h.headers {
		cb(n, v)
	}

}

// validate header bytes like \r, \n, etc
func validateHeaderBytes(data []byte) error {
	for i := 0; i < len(data); i++ {
		b := data[i]

		if b == '\n' {
			if i == 0 || data[i-1] != '\r' {
				return ERROR_BAD_HEADER
			}
			continue
		}

		if b == '\r' {
			if i+1 >= len(data) {
				return nil // incomplete
			}
			if data[i+1] != '\n' {
				return ERROR_BAD_HEADER
			}
			i++
			continue
		}

		if b < 0x20 || b == 0x7f {
			return ERROR_BAD_HEADER
		}
	}
	return nil
}

// check if name field has valid token or charaters
func isValidToken(str []byte) bool {

	for _, ch := range str {

		found := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {

			found = true
		}

		switch ch {

		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':

			found = true
		}

		if !found {

			return false
		}

	}
	return true
}

func parseHeaderLine(fieldLine []byte) (string, string, error) {

	// split name and value field separated by colon
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid Header")
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1]) // Trims all white space in value field

	// check if name field has whitespace if yes it is invalid
	if bytes.HasSuffix(name, []byte(" ")) {

		return "", "", fmt.Errorf("Invalid Header (invalid name field)\n")
	}

	return string(name), string(value), nil
}

func (h *Headers) Parse(data []byte) (int, bool, error) {

	if err := validateHeaderBytes(data); err != nil {

		return 0, false, err
	}

	read := 0

	done := false

	for {

		index := bytes.Index(data[read:], crlf)
		if index == -1 {

			break
		}

		// empty header
		if index == 0 {

			done = true
			read += len(crlf)
			break
		}

		name, value, err := parseHeaderLine(data[read : read+index])
		if err != nil {

			return 0, false, err
		}

		if !isValidToken([]byte(name)) {

			return 0, false, fmt.Errorf("Token Invalid...")
		}

		read += index + len(crlf)

		h.Set(name, value)

	}

	return read, done, nil

}
