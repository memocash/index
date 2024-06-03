package server

import (
	"fmt"
	"github.com/memocash/index/ref/config"
	"net"
	"net/http"
)

type Server struct {
	Port     uint
	server   http.Server
	listener net.Listener
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return fmt.Errorf("error starting admin server; %w", err)
	}
	// Serve always returns an error
	return fmt.Errorf("error serving admin server; %w", s.Serve())
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", GetIndexHandler())
	mux.HandleFunc("/graphql", GetGraphQLHandler())
	s.server = http.Server{Handler: mux}
	var err error
	if s.listener, err = net.Listen("tcp", config.GetHost(s.Port)); err != nil {
		return fmt.Errorf("failed to listen admin server; %w", err)
	}
	return nil
}

func (s *Server) Serve() error {
	if err := s.server.Serve(s.listener); err != nil {
		return fmt.Errorf("error listening and serving admin server; %w", err)
	}
	return fmt.Errorf("error admin server disconnected")
}

func NewServer() *Server {
	return &Server{
		Port: config.GetGraphQLPort(),
	}
}
