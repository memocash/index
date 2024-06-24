package server

import (
	"bufio"
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

func (s *Server) GetHost() string {
	return config.GetHost(s.Port)
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
	if s.listener, err = net.Listen("tcp", s.GetHost()); err != nil {
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

type sizeWriter struct {
	httpWriter http.ResponseWriter
	totalSize  int
}

func (w *sizeWriter) Write(b []byte) (int, error) {
	size, err := w.httpWriter.Write(b)
	w.totalSize += size
	return size, err
}

func (w *sizeWriter) Header() http.Header {
	return w.httpWriter.Header()
}

func (w *sizeWriter) WriteHeader(statusCode int) {
	w.httpWriter.WriteHeader(statusCode)
}

func (w *sizeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hi, ok := w.httpWriter.(http.Hijacker); ok {
		return hi.Hijack()
	}
	return nil, nil, fmt.Errorf("http.Hijacker not implemented")
}
