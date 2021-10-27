package server

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/admin/server/graphql"
	"github.com/memocash/server/admin/server/network"
	node2 "github.com/memocash/server/admin/server/node"
	"github.com/memocash/server/node"
	"github.com/memocash/server/ref/config"
	"net/http"
)

type Server struct {
	Nodes *node.Group
	Port  uint
}

var routes = admin.Routes([]admin.Route{
	indexRoute,
},
	network.GetRoutes(),
	node2.GetRoutes(),
	//graphql.GetRoutes(),
)

func (s *Server) Run() error {
	mux := http.NewServeMux()
	for _, tempRoute := range routes {
		route := tempRoute
		mux.HandleFunc(route.Pattern, func(w http.ResponseWriter, r *http.Request) {
			route.Handler(admin.Response{
				Writer:    w,
				Request:   r,
				NodeGroup: s.Nodes,
				Route:     route,
			})
			jlog.Logf("Processed admin request: %s\n", route.Pattern)
		})
	}
	graphqlHandler, err := graphql.GetHandler()
	if err != nil {
		return jerr.Get("error getting graphql handler", err)
	}
	mux.Handle(admin.UrlGraphql, graphqlHandler)
	server := http.Server{
		Addr:    config.GetHost(s.Port),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		return jerr.Get("error listening and serving admin server", err)
	}
	return nil
}

func NewServer(group *node.Group) *Server {
	return &Server{
		Nodes: group,
		Port:  config.GetAdminPort(),
	}
}
