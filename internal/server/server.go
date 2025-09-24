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
	state    string
	listener net.Listener
}

func NewServer() *Server {
	return &Server{}
}

func Serve(port uint16) (*Server, error) {
	server := NewServer()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server.listener = listener

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

	request.PrintRequest()

	// Response
	response.WriteStatusLine(c, response.OK)
	response.WriteHeaders(c, *response.GetDefaultHeaders(0))

}
