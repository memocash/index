package admin

import (
	"github.com/memocash/server/db/item"
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
}

type NodeDisconnectRequest struct {
	NodeId string
}

type NodeConnectRequest struct {
	Ip   []byte
	Port uint16
}

type NodeHistoryRequest struct {
	SuccessOnly bool
}

type NodeHistoryResponse struct {
	Connections []*item.PeerConnection
}

type NodeFoundPeersRequest struct {
	Ip   []byte
	Port uint16
}

type NodeFoundPeersResponse struct {
	FoundPeers []*item.FoundPeer
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
)
