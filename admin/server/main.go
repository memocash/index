package server

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/admin/server/graphql"
	"github.com/memocash/index/admin/server/network"
	node2 "github.com/memocash/index/admin/server/node"
	"github.com/memocash/index/admin/server/topic"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/config"
	"net"
	"net/http"
)

type Server struct {
	Nodes    *node.Group
	Port     uint
	server   http.Server
	listener net.Listener
}

var routes = admin.Routes([]admin.Route{
	indexRoute,
},
	network.GetRoutes(),
	node2.GetRoutes(),
	topic.GetRoutes(),
	//graphql.GetRoutes(),
)

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return jerr.Get("error starting admin server", err)
	}
	// Serve always returns an error
	return jerr.Get("error serving admin server", s.Serve())
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	for _, tempRoute := range routes {
		route := tempRoute
		mux.HandleFunc(route.Pattern, getHandler(func(w http.ResponseWriter, r *http.Request) {
			route.Handler(admin.Response{
				Writer:    w,
				Request:   r,
				NodeGroup: s.Nodes,
				Route:     route,
			})
		}, false))
	}
	graphqlHandler, err := graphql.GetHandler()
	if err != nil {
		return jerr.Get("error getting graphql handler", err)
	}
	mux.HandleFunc(admin.UrlGraphql, getHandler(graphqlHandler.ServeHTTP, true))
	s.server = http.Server{Handler: mux}
	if s.listener, err = net.Listen("tcp", config.GetHost(s.Port)); err != nil {
		return jerr.Get("failed to listen admin server", err)
	}
	return nil
}

func (s *Server) Serve() error {
	if err := s.server.Serve(s.listener); err != nil {
		return jerr.Get("error listening and serving admin server", err)
	}
	return jerr.New("error admin server disconnected")
}

func NewServer(group *node.Group) *Server {
	return &Server{
		Nodes: group,
		Port:  config.GetAdminPort(),
	}
}

func getHandler(handler func(http.ResponseWriter, *http.Request), corsAllOrigins bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if corsAllOrigins {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", "*")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Server")
		}
		handler(w, r)
		jlog.Logf("Processed admin request: %s\n", r.URL)
	}
}
