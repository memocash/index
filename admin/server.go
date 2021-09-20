package admin

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/config"
	"net/http"
)

type Server struct {
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo Admin 0.1")
	})
	server := http.Server{
		Addr:    config.GetHost(config.GetAdminPort()),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		return jerr.Get("error listening and serving admin server", err)
	}
	return nil
}

func NewServer() *Server {
	return &Server{}
}
