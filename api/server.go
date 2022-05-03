package api

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
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
		return jerr.Get("error starting api server", err)
	}
	// Serve always returns an error
	return jerr.Get("error serving api server", s.Serve())
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo API 0.1")
	})
	s.server = http.Server{Handler: mux}
	var err error
	if s.listener, err = net.Listen("tcp", config.GetHost(s.Port)); err != nil {
		return jerr.Get("failed to listen api server", err)
	}
	return nil
}

func (s *Server) Serve() error {
	if err := s.server.Serve(s.listener); err != nil {
		return jerr.Get("error listening and serving api server", err)
	}
	return jerr.New("error api server disconnected")
}

func NewServer() *Server {
	return &Server{
		Port: config.GetApiPort(),
	}
}
