package admin

import "github.com/memocash/server/db/item"

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
