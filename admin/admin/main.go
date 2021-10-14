package admin

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/node"
	"net/http"
)

type Route struct {
	Pattern string
	Handler func(Response)
}

func Routes(routes ...[]Route) []Route {
	var allRoutes []Route
	for _, routeGroup := range routes {
		allRoutes = append(allRoutes, routeGroup...)
	}
	return allRoutes
}

type Response struct {
	Writer    http.ResponseWriter
	Request   *http.Request
	NodeGroup *node.Group
	Route     Route
}

func (r Response) Error(err error) {
	jerr.Getf(err, "error with request: %s", r.Route.Pattern).Print()
}

const (
	UrlIndex               = "/"
	UrlNodeGetAddrs        = "/node/get_addrs"
	UrlNodeConnect         = "/node/connect"
	UrlNodeConnectDefault  = "/node/connect_default"
	UrlNodeConnectNext     = "/node/connect_next"
	UrlNodeListConnections = "/node/list_connections"
	UrlNodeDisconnect      = "/node/disconnect"
	UrlNodeHistory         = "/node/history"
	UrlNodeLoopingEnable   = "/node/looping_enable"
	UrlNodeLoopingDisable  = "/node/looping_disable"
	UrlNodeFoundPeers      = "/node/found_peers"
	UrlNodePeers           = "/node/peers"
	UrlNodePeerReport      = "/node/peer_report"
)
