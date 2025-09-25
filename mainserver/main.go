package main

import (
	"fmt"
	"http/internal/headers"
	"http/internal/request"
	"http/internal/response"
	"http/internal/server"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const port = 3030

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	slog.Info("Server started on", "port", port)

	// Common pattern for gracefully shutting down a server.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println()
	slog.Info("Server gracefully stopped")
}

func handler(res *response.Response, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/not":
		return &server.HandlerError{StatusCode: &response.NOT_FOUND, Message: []byte("Nothing to say :(")}
	case "/bad":
		return &server.HandlerError{StatusCode: &response.BAD_REQUEST}
	case "/server-error":
		return &server.HandlerError{StatusCode: &response.INTERNAL_SERVER_ERROR, Message: []byte("My bad :|")}
	case "/chunked":
		req.PrintRequest()

		// Step 1: write headers
		heders := headers.NewHeaders()
		heders.Set(headers.CONTENT_TYPE, "text/plain")
		heders.Set("Transfer-Encoding", "chunked")
		res.Write(&response.OK, heders, nil)

		// Step 1: write chunk
		bigData := generateBigData(100 * 1024 * 1024) // 100MB
		cunckedSize := 100
		size := fmt.Appendf(nil, "%x\r\n", cunckedSize)
		for i := 0; i < len(bigData); i += cunckedSize {
			res.WriteChunkedBody(size, bigData[i:i+cunckedSize])
		}
		res.WriteChunkedBodyDone()

	default:
		req.PrintRequest()
		body := "Good!\n"
		res.Write(&response.OK, nil, []byte(body))
	}
	return nil
}

// NETCAT command to test chunk data
// printf "POST /chunked HTTP/1.1\r\nHost: localhost:3030\r\nContent-Type: application/json\r\nContent-Length: 55\r\n\r\n{\"type\": \"dark mode\", \"size\": \"medium\",\"billy\":\"ballo\"}" | nc localhost 3030
func generateBigData(size int) []byte {
	data := make([]byte, size)
	pattern := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range size {
		data[i] = pattern[i%len(pattern)]
	}
	return data
}
