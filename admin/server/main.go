package server

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/admin/server/graphql"
	"github.com/memocash/index/admin/server/network"
	node2 "github.com/memocash/index/admin/server/node"
	"github.com/memocash/index/node"
	"github.com/memocash/index/ref/config"
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
		mux.HandleFunc(route.Pattern, getHandler(func(w http.ResponseWriter, r *http.Request) {
			route.Handler(admin.Response{
				Writer:    w,
				Request:   r,
				NodeGroup: s.Nodes,
				Route:     route,
			})
		}))
	}
	graphqlHandler, err := graphql.GetHandler()
	if err != nil {
		return jerr.Get("error getting graphql handler", err)
	}
	mux.HandleFunc(admin.UrlGraphql, getHandler(graphqlHandler.ServeHTTP))
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

func getHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
		jlog.Logf("Processed admin request: %s\n", r.URL)
	}
}
