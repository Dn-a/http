package server

import (
	"fmt"
	"http/internal/request"
	"http/internal/response"
	"log"
	"log/slog"
	"net"
)

type Server struct {
	closed   bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

type Handler func(res *response.Response, req *request.Request) *HandlerError

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
		}
		if err != nil {
			log.Fatal("Connection Error: ", err)
		}

		slog.Info("New clinet", "addr", conn.RemoteAddr())

		go s.handle(conn)
	}
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()

	request, err := request.RequestFromReader(c)
	if err != nil {
		log.Fatal("Request error: ", err)
	}

	resp := &response.Response{Writer: c}

	if hErr := s.handler(resp, request); hErr != nil {
		request.PrintRequest()
		resp.Status(&hErr.StatusCode)
		resp.Body(hErr.Message)
		resp.Write()
	}
}
