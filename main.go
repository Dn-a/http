package main

import (
	"bytes"
	"fmt"
	"http/internal/request"
	"io"
	"log"
	"net"
	"os"
)

const SERVER_PORT = 3030
const BUFFER_SIZE = 30
const TIMEOUT = 10

func main() {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", SERVER_PORT))

	if err != nil {
		log.Fatal("Listener error: ", err)
	}
	fmt.Printf("Server started on %d\n\n", SERVER_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Connection Error: ", err)
		}

		fmt.Printf("New connection: %v", conn.RemoteAddr())

		// Anonymous function executed as goroutine - enables concurrent handling
		// of multiple client connections.
		go func(c net.Conn) {
			defer c.Close()
			r, err := request.RequestFromReader(conn)
			if err != nil {
				log.Fatal("Request error: ", err)
			}

			fmt.Println("Request Line:")
			fmt.Printf("- Method: %s\n", r.RequestLine.Method)
			fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
			fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
			fmt.Println("Headers:")
			r.Headers.ForEach(func(k, v string) {
				fmt.Printf("- %s: %s\n", k, v)
			})

		}(conn)

		// ANSI escape sequence to clear terminal
		fmt.Print("\033[2J\033[H")
	}

}

// to be removed
func getFile(fName string) *os.File {
	f, err := os.Open(fName)
	if err != nil {
		log.Fatal("error", "error", err)
	}
	return f
}

// to be removed
func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)

		buffer := make([]byte, BUFFER_SIZE)
		lnIdx := -1
		for {
			_, err := f.Read(buffer)
			if err != nil {
				break
			}

			if lnIdx = bytes.IndexByte(buffer, '\n'); lnIdx != -1 {
				out <- string(buffer[:lnIdx+1])
			}

			if len(buffer[lnIdx+1:]) > 0 {
				out <- string(buffer[lnIdx+1:])
			}
			buffer = make([]byte, BUFFER_SIZE)
		}
	}()
	return out
}
