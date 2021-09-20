package admin

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/node"
	"github.com/memocash/server/ref/config"
	"net/http"
)

const (
	UrlIndex        = "/"
	UrlNodeGetAddrs = "/node/get_addrs"
)

type Server struct {
	Node *node.Server
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(UrlIndex, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo Admin 0.1")
	})
	mux.HandleFunc(UrlNodeGetAddrs, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node get addrs request")
		s.Node.GetAddr()
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

func NewServer(nodeServer *node.Server) *Server {
	return &Server{
		Node: nodeServer,
	}
}
