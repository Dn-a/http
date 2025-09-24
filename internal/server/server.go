package server

import (
	"fmt"
	"http/internal/request"
	"http/internal/response"
	"io"
	"log"
	"log/slog"
	"net"
)

type Server struct {
	state    string
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func NewServer() *Server {
	return &Server{}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	server := NewServer()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server.listener = listener
	server.handler = handler

	go server.listen()

	return server, err
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
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

	if hErr := s.handler(c, request); hErr != nil {
		request.PrintRequest()
		response.WriteStatusLine(c, hErr.StatusCode)
		response.WriteHeaders(c, response.GetDefaultHeaders(len(hErr.Message)))
		response.WriteBody(c, hErr.Message)
	}
}
