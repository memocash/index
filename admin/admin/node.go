package admin

import (
	"github.com/memocash/server/db/item"
	"time"
)

type NodeDisconnectRequest struct {
	NodeId string
}

type NodeConnectRequest struct {
	Ip   []byte
	Port uint16
}

type NodeHistoryRequest struct {
	SuccessOnly bool
	Ip          string
	Port        uint32
}

type Connection struct {
	Ip     string
	Port   uint16
	Time   time.Time
	Status item.PeerConnectionStatus
}

type NodeHistoryResponse struct {
	Connections []Connection
}

type NodeFoundPeersRequest struct {
	Ip   []byte
	Port uint16
}

type NodeFoundPeersResponse struct {
	FoundPeers []*item.FoundPeer
}

type NodePeersRequest struct {
	Page uint
}

type NodePeersResponse struct {
	Peers []*Peer
}

type Peer struct {
	Ip       string
	Port     uint16
	Services uint64
}
