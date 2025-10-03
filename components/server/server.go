package server

import (
	"fmt"
	"http/components/headers"
	"http/components/request"
	"http/components/response"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
)

type Handler func(res *response.Response, req *request.Request) *HandlerError
type HandlerError struct {
	StatusCode *response.StatusCode
	Message    []byte
}

func (he *HandlerError) Write(res *response.Response) {
	var (
		body           []byte
		currentHeaders *headers.Headers
		err            error
	)
	if he.Message == nil {
		if body, err = os.ReadFile(filepath.Join("internal", "error", fmt.Sprintf("%d.html", he.StatusCode.Code))); err != nil {
			fmt.Println(err)
			body = []byte("Unprocessed error.")
		} else {
			currentHeaders = response.GetDefaultHeaders(len(body))
			currentHeaders.Set(headers.CONTENT_TYPE, "text/html")
		}
	} else {
		body = fmt.Appendf(nil, "{\"statusCode\":%d, \"errorMessage\":\"%s\"}\n", he.StatusCode.Code, he.Message)
	}
	res.Write(he.StatusCode, currentHeaders, body)
}

type Server struct {
	closed   bool
	listener net.Listener
	handler  Handler
}

func Serve(port uint16, handler Handler) (*Server, error) {
	server := &Server{closed: false, handler: handler}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server.listener = listener

	go server.listen()

	return server, err
}

func (s *Server) Close() error {
	s.closed = true
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.closed {
			fmt.Println("Server closed")
			break
		} else if err != nil {
			log.Fatal("Connection Error: ", err)
			continue
		}

		slog.Info("New clinet", "addr", conn.RemoteAddr())

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	resp := &response.Response{Writer: conn}
	request, err := request.RequestFromReader(conn)

	if err != nil {
		fmt.Printf("Request error: %v", err)
		hErr := &HandlerError{
			StatusCode: &response.BAD_REQUEST,
			Message:    []byte(err.Error()),
		}
		hErr.Write(resp)
		return
	}

	hErr := s.handler(resp, request)

	if hErr != nil {
		request.PrintRequest()
		hErr.Write(resp)
	}
}
