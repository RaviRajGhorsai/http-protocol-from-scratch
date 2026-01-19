package main

import (
	"crypto/sha256"
	"fmt"
	headers "htttpfromtcp/internal/header"
	"htttpfromtcp/internal/request"
	"htttpfromtcp/internal/response"
	"htttpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func respond200() []byte {

	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

}

func respond400() []byte {

	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {

	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func toStr(bytes []byte) string {
	out := ""

	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}

	return out

}

func logging(req *request.Request, h *headers.Headers) {

	fmt.Printf("Request line:\n")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

	fmt.Println("Headers:")
	req.Headers.ForEach(func(n, v string) {

		fmt.Printf("- %s: %s\n", n, v)

	})

	fmt.Println("Body: ")
	fmt.Printf("%s", req.Body)

	fmt.Printf("\nTrailers: ")
	fmt.Printf("%s", h.Get("trailer"))

}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {

		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK

		// filter the request target and respond accordingly (later we can add find functionality accordingly)
		if req.RequestLine.RequestTarget == "/yourproblem" {

			body = respond400()
			status = response.StatusBadRequest

		} else if req.RequestLine.RequestTarget == "/myproblem" {

			body = respond500()
			status = response.StatusInternalServerError

		} else if req.RequestLine.RequestTarget == "/video" {

			videoData, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				log.Printf("Error: %v\n", err)
				return
			}

			h.Replace("Content-Length", fmt.Sprintf("%d", len(videoData)))
			h.Replace("content-type", "video/mp4")

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(videoData)

			// Chunked Encoding,  using httpbin that sends chunked data
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {

			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])

			if err != nil {
				log.Printf("Error 1: %v\n", err)
				body = respond400()
				status = response.StatusBadRequest

			} else {
				defer res.Body.Close()

				w.WriteStatusLine(response.StatusOK)

				h.Delete("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("Content-Type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")

				w.WriteHeaders(h)

				fullBody := []byte{}

				for {

					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {

						log.Printf("Error 2: %s\n", err)
						break
					}

					fullBody = append(fullBody, data[:n]...)

					_, e := w.WriteChunkedBody(data[:n])
					if e != nil {

						log.Printf("Error 3: %s\n", err)
						break
					}
				}

				w.WriteChunkedBodyDone()

				trailer := headers.NewHeaders()

				out := sha256.Sum256(fullBody)

				trailer.Set("X-Content-SHA256", toStr(out[:]))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

				w.WriteBody([]byte("0\r\n"))
				w.WriteTrailers(trailer)

				// Loggings only

				logging(req, h)

				return
			}
		}

		// replace content length with actual length of body
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")

		// these response must be written in order first statusline, header then body
		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(body)

		// Loggings only

		logging(req, h)

	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
