package api

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"net/http"
)

type Server struct {
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo API 0.1")
	})
	server := http.Server{
		Addr:    ":10000",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		return jerr.Get("error listening and serving server", err)
	}
	return nil
}

func NewServer() *Server {
	return &Server{}
}
