package main

import (
	"fmt"
	"http/internal/server"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const port = 3030

func main() {
	server, err := server.Serve(port)
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
