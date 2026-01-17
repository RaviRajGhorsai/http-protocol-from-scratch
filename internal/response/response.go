package response

import (
	"fmt"
	headers "htttpfromtcp/internal/header"
	"io"
)

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// define custom writer struct to write stuffs
type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}

}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "closed")
	h.Set("Content-Type", "text/html")

	return h
}



// writes status line to the response
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		_, err := w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err

	case StatusBadRequest:
		_, err := w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err

	case StatusInternalServerError:
		_, err := w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	}

	return fmt.Errorf("Unknown Status")

}


// writes headers to the response
func (w *Writer) WriteHeaders(h *headers.Headers) error {
	b := []byte{}

	h.ForEach(func(n, v string) {

		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})

	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)

	return err

}

func (w *Writer) WriteBody(p []byte) (int, error) {

	n, err := w.writer.Write(p)

	return n, err
}
