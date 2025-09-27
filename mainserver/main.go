package main

import (
	"crypto/sha256"
	"fmt"
	"http/internal/headers"
	"http/internal/request"
	"http/internal/response"
	"http/internal/server"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

		// Step 2: write chunk
		size := 1024 // Byte
		bigData := generateBigData(size)
		cunckedSize := 100
		var end int
		for i := 0; i < len(bigData); i += cunckedSize {
			end = i + cunckedSize
			if end > len(bigData) {
				end = len(bigData)
			}
			res.WriteChunkedBody(bigData[i:end])
		}
		// Step 3: close body
		res.WriteChunkedBodyDone()
	case "/chunked-trailer":
		req.PrintRequest()

		// Step 1: write headers
		h := headers.NewHeaders()
		h.Set(headers.CONTENT_TYPE, "text/plain")
		h.Set("Transfer-Encoding", "chunked")
		h.Set("Trailer", "x-content-sha256, x-content-length")
		res.Write(&response.OK, h, nil)

		// Step 2: write chunk
		size := 1024 * 1024 // Byte
		bigData := generateBigData(size)
		cunckedSize := 100
		var end int
		for i := 0; i < len(bigData); i += cunckedSize {
			end = i + cunckedSize
			if end > len(bigData) {
				end = len(bigData)
			}
			res.WriteChunkedBody(bigData[i:end])
		}

		// Step 3: write trailer
		trailer := headers.NewHeaders()
		trailer.Set("X-Content-SHA256", sha256Encode(sha256.Sum256(bigData)))
		trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(bigData)))
		res.WriteTrailers(trailer)

	case "/binary":
		req.PrintRequest()

		// STEP 1: write Headers
		h := headers.NewHeaders()
		h.Set(headers.CONTENT_TYPE, "video/mp4")
		h.Set("Transfer-Encoding", "chunked")
		h.Set("Trailer", "x-content-length")
		res.Write(&response.OK, h, nil)

		// Step 2: write video
		file, err := os.Open(filepath.Join("assets", "test.mp4"))
		defer file.Close()
		if err != nil {
			fmt.Println(err)
		}
		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println(err)
			break
		}

		size := fileInfo.Size()
		buffer := make([]byte, 1024)
		for {
			_, err := file.Read(buffer)
			if err != nil {
				break
			}
			res.WriteChunkedBody(buffer)
		}

		// Step 3: write trailer
		trailer := headers.NewHeaders()
		trailer.Set("X-Content-Length", fmt.Sprintf("%d", size))
		res.WriteTrailers(trailer)

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

func sha256Encode(bytes [32]byte) string {
	var b strings.Builder
	for _, by := range bytes {
		b.WriteString(fmt.Sprintf("%02x", by))
	}
	return b.String()
}
