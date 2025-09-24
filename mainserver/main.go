package main

import (
	"fmt"
	"http/internal/request"
	"http/internal/response"
	"http/internal/server"
	"io"
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

func handler(w io.Writer, r *request.Request) *server.HandlerError {
	switch r.RequestLine.RequestTarget {
	case "/bad":
		return &server.HandlerError{StatusCode: response.NOT_FOUND, Message: []byte("Nothing to say :(")}
	case "/server-error":
		return &server.HandlerError{StatusCode: response.INTERNAL_SERVER_ERROR, Message: []byte("My bad :|")}
	default:
		message := "Good!"
		r.PrintRequest()
		response.WriteStatusLine(w, response.OK)
		response.WriteHeaders(w, response.GetDefaultHeaders(len(message)))
		response.WriteBody(w, []byte(message))
	}
	return nil
}
