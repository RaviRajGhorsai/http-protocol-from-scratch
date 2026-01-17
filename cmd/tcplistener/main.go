package main

import (
	"fmt"
	"htttpfromtcp/internal/request"
	"log"
	"net"
)

func main() {

	// listener on tcp port 42069
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {

		fmt.Println("error listening on port 42069...")
		return
	}

	// defer will make sure when program exits it closes listener
	defer listener.Close()

	fmt.Println("Server Listening in port 42069")

	for {

		conn, err := listener.Accept()

		if err != nil {

			fmt.Println("Error on establishing connection....")
			continue
		}

		fmt.Println("Connection has been established...")

		// prints the messages that is read as 8 byte chunk
		r, err := request.RequestFromReader(conn)
		if err != nil {

			log.Fatal("error", "error", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		r.Headers.ForEach(func(n, v string) {

			fmt.Printf("- %s: %s\n", n, v)	

		})

	}

}
