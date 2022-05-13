package admin

import (
	"github.com/memocash/index/db/item"
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
	Port        uint32 `json:",string"`
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
	Page   uint
	Filter string
}

type NodePeersResponse struct {
	Peers []*Peer
}

type Peer struct {
	Ip       string
	Port     uint16
	Services uint64
	Time     time.Time
	Status   item.PeerConnectionStatus
}

type NodePeerReportResponse struct {
	TotalPeers     uint64
	PeersAttempted uint64
	TotalAttempts  uint64
	PeersConnected uint64
	PeersFailed    uint64
}

type Topic struct {
	Name string
}

type TopicListResponse struct {
	Topics []Topic
}

type TopicViewRequest struct {
	Topic string
	Start string
	Shard int
}

type TopicItem struct {
	Topic   string
	Uid     string
	Message string
	Shard   uint
	Props   map[string]interface{}
}

type TopicViewResponse struct {
	Name  string
	Items []TopicItem
}

type TopicItemRequest struct {
	Topic string
	Shard uint
	Uid   string
}

type TopicItemResponse struct {
	Item TopicItem
}
