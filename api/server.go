package api

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/config"
	"net/http"
)

type Server struct {
	Port uint
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo API 0.1")
	})
	server := http.Server{
		Addr:    config.GetHost(s.Port),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		return jerr.Get("error listening and serving api server", err)
	}
	return nil
}

func NewServer() *Server {
	return &Server{
		Port: config.GetApiPort(),
	}
}
