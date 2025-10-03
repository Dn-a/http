# HTTP Server Implementation in Go

A minimalist HTTP/1.1 server implementation in Go, developed from scratch **for educational purposes only**. 
This project is not intended to replace production-ready HTTP servers or frameworks, but rather to explore and understand the underlying mechanics of HTTP protocol handling. 
(For production applications, use established frameworks like Go's `net/http` standard library or frameworks like Gin, Echo, or Fiber)
This project demonstrates low-level HTTP protocol implementation concepts including request parsing, response generation, chunked transfer encoding, and trailer headers. It was created to deepen understanding of how HTTP servers work under the hood.

## Technical Notes

**HTTP/1.1 Compliance:**
- CRLF (`\r\n`) line terminators
- Request line format: `METHOD TARGET HTTP/VERSION`
- Header format: `Field-Name: Field-Value`
- Double CRLF separates headers from body

**Buffer Management:**
Request parser uses sliding window technique to handle partial reads efficiently.

**Concurrency:**
Each client connection is handled in a separate goroutine, allowing multiple simultaneous requests.

## Limitations

This is an educational implementation and **should not be used in production**. It lacks:
- TLS/HTTPS support
- Request timeout handling
- Connection pooling
- HTTP/2 or HTTP/3 support
- Comprehensive error recovery
- Security
- Performance optimizations

## Learning Resources

To understand this implementation, study these concepts:
- HTTP/1.1 specification (RFC 7230-7235)
- TCP socket programming in Go
- Chunked transfer encoding (RFC 7230 Section 4.1)
- Trailer headers (RFC 7230 Section 4.1.2)
- Go's `net` package and goroutines

## Main Components

### Server (`server.go`)

The server module handles TCP connections and coordinates request/response processing.

**Key features:**
- Listens on TCP port 3030(default)
- Accepts incoming connections and spawns goroutines for concurrent handling
- Shutdown with signal handling (SIGINT, SIGTERM)
- Custom error handling with status codes

**Example flow:**
```
Client connects → server.listen() accepts → server.handle() processes
→ Parses request → Executes handler → Writes response → Closes connection
```

### Request Parser (`request.go`)

Implements stateful HTTP request parsing with a finite state machine.

**States:**
- `RequestInit` - Parse request line (Method, Target, HTTP Version) with CRLF delimiters
- `RequestHeaders` - Parse headers with CRLF delimiters
- `RequestBody` - Read body based on Content-Length
- `RequestDone` - Complete parsing
- `RequestError` - Handle parsing errors

**Example request line parsing:**
```
Input: "GET /chunked HTTP/1.1\r\n"
Output: Method="GET", RequestTarget="/chunked", HttpVersion="1.1"
```

**Key features:**
- Dynamic buffer management (1024 bytes initial capacity)
- Handles incomplete data reads
- Validates Content-Length against actual body size
- CRLF delimiter detection for HTTP/1.1 compliance

### Response Writer (`response.go`)

Generates HTTP responses with support for standard and chunked transfer encoding.

**Methods:**
- `Write()` - Standard response with Content-Length
- `WriteChunkedBody()` - Stream data in chunks with hex-encoded sizes
- `WriteTrailers()` - Append trailer headers after chunked body
- `WriteChunkedBodyDone()` - Signal end of chunked transfer

**Example chunked encoding:**
```
Response header: Transfer-Encoding: chunked
Chunk format: 64\r\n[100 bytes of data]\r\n
End signal: 0\r\n\r\n
```

**Example with trailers:**
```
Transfer-Encoding: chunked
Trailer: x-content-sha256

64\r\n
[data chunk]\r\n
0\r\n
x-content-sha256: [computed hash]\r\n
\r\n
```

### Headers Management (`headers.go`)

Provides header parsing, validation, and storage.

**Features:**
- Case-insensitive header names (stored lowercase)
- Token validation for field names (RFC 7230 compliance)
- Batch parsing with `ParseAll()` method
- CRLF detection for header boundaries

**Valid token characters:**
```
Alphanumeric: a-z, A-Z, 0-9
Special: ! # $ % & ' * + - . ^ _ ` | ~
```

## Route Examples

The `main.go` file defines several demonstration endpoints:

### `/chunked` - Basic Chunked Transfer
Generates 1KB of data sent in 100-byte chunks.

**Response:**
```
HTTP/1.1 200 OK
Content-Type: text/plain
Transfer-Encoding: chunked

64\r\n
ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRST\r\n
64\r\n
[more chunks...]\r\n
0\r\n\r\n
```

### `/chunked-trailer` - Chunked with Trailers
Sends 1MB of data with SHA-256 checksum in trailer headers.

**Response includes:**
```
Trailer: x-content-sha256, x-content-length
[chunked body]
x-content-sha256: [hash] (hex)
x-content-length: 1048576
```

### `/binary` - File Streaming
Streams a video file (test.mp4) using chunked encoding with file size in trailers.

### Error Routes
- `/not` → 404 Not Found
- `/bad` → 400 Bad Request
- `/server-error` → 500 Internal Server Error

## Running the Server

**Start:**
```
go run main.go
**Output:**
Server started on port 3030

**Test with curl:**
curl --raw http://localhost:3030/chunked

**Test with netcat:**
printf "POST /chunked HTTP/1.1\r\nHost: localhost:3030\r\nContent-Type: application/json\r\nContent-Length: 55\r\n\r\n{"type": "dark mode", "size": "medium","billy":"ballo"}" | nc localhost 3030
```

## Testing

The project includes unit tests for core components:
- `request_test.go` - Request parsing validation
- `response_test.go` - Response generation verification  
- `headers_test.go` - Header parsing and validation
