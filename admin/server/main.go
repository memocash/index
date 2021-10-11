package server

import (
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/node"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net"
	"net/http"
)

type Server struct {
	Nodes *node.Group
	Port  uint
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(admin.UrlIndex, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo Admin 0.1")
	})
	mux.HandleFunc(admin.UrlNodeGetAddrs, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node get addrs request")
		for _, serverNode := range s.Nodes.Nodes {
			serverNode.GetAddr()
		}
	})
	mux.HandleFunc(admin.UrlNodeConnect, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node connect")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jerr.Get("error reading node connect body", err).Print()
			return
		}
		var connectRequest = new(admin.NodeConnectRequest)
		if err := json.Unmarshal(body, connectRequest); err != nil {
			jerr.Get("error unmarshalling node connect request", err).Print()
			return
		}
		s.Nodes.AddNode(connectRequest.Ip, connectRequest.Port)
	})
	mux.HandleFunc(admin.UrlNodeConnectDefault, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node connect default")
		s.Nodes.AddDefaultNode()
	})
	mux.HandleFunc(admin.UrlNodeConnectNext, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node connect next")
		s.Nodes.AddNextNode()
	})
	mux.HandleFunc(admin.UrlNodeLoopingEnable, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node looping enable")
		if s.Nodes.Looping {
			return
		}
		s.Nodes.Looping = true
		if !s.Nodes.HasActive() {
			s.Nodes.AddNextNode()
		}
	})
	mux.HandleFunc(admin.UrlNodeLoopingDisable, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node looping disabled")
		s.Nodes.Looping = false
	})
	mux.HandleFunc(admin.UrlNodeListConnections, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node list connections")
		for id, serverNode := range s.Nodes.Nodes {
			fmt.Fprintf(w, "%s - %s:%d (%t)\n", id, net.IP(serverNode.Ip), serverNode.Port,
				serverNode.Peer != nil && serverNode.Peer.Connected())
		}
	})
	mux.HandleFunc(admin.UrlNodeHistory, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node list history")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jerr.Get("error reading node history body", err).Print()
			return
		}
		var historyRequest = new(admin.NodeHistoryRequest)
		if err := json.Unmarshal(body, historyRequest); err != nil {
			jerr.Get("error unmarshalling node history request", err).Print()
			return
		}
		var foundPeerConnections []*item.PeerConnection
		var startId []byte
		var shard uint32
	PeerConnectionsLoop:
		for {
			peerConnections, err := item.GetPeerConnections(shard, startId)
			if err != nil {
				jerr.Get("fatal error getting peer connections", err).Fatal()
			}
			for _, peerConnection := range peerConnections {
				if !historyRequest.SuccessOnly || peerConnection.Status == item.PeerConnectionStatusSuccess {
					foundPeerConnections = append(foundPeerConnections, peerConnection)
					if len(foundPeerConnections) >= client.LargeLimit {
						break PeerConnectionsLoop
					}
				}
			}
			if len(peerConnections) < client.LargeLimit {
				shard++
				if shard >= config.GetTotalShards() {
					break
				}
			}
		}
		var historyResponse = &admin.NodeHistoryResponse{
			Connections: foundPeerConnections,
		}
		historyResponseData, err := json.Marshal(historyResponse)
		if err != nil {
			jerr.Get("error marshalling history response data", err).Print()
			return
		}
		if _, err = w.Write(historyResponseData); err != nil {
			jerr.Get("error writing history response data", err).Print()
			return
		}
	})
	mux.HandleFunc(admin.UrlNodeFoundPeers, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node found peers request")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jerr.Get("error reading node found peers request", err).Print()
			return
		}
		var request = new(admin.NodeFoundPeersRequest)
		if err := json.Unmarshal(body, request); err != nil {
			jerr.Get("error unmarshalling found peers request", err).Print()
			return
		}
		var foundFoundPeers []*item.FoundPeer
		var startId []byte
		var shard uint32
	FoundPeersLoop:
		for {
			foundPeers, err := item.GetFoundPeers(shard, startId, request.Ip, request.Port)
			if err != nil {
				jerr.Get("fatal error getting found peers", err).Fatal()
			}
			for _, foundPeer := range foundPeers {
				foundFoundPeers = append(foundFoundPeers, foundPeer)
				if len(foundFoundPeers) >= client.LargeLimit {
					break FoundPeersLoop
				}
			}
			if len(foundPeers) < client.LargeLimit {
				shard++
				if shard >= config.GetTotalShards() {
					break
				}
			}
		}
		var foundPeersResponse = &admin.NodeFoundPeersResponse{
			FoundPeers: foundFoundPeers,
		}
		foundPeersResponseData, err := json.Marshal(foundPeersResponse)
		if err != nil {
			jerr.Get("error marshalling found peers response data", err).Print()
			return
		}
		if _, err = w.Write(foundPeersResponseData); err != nil {
			jerr.Get("error writing found peers response data", err).Print()
			return
		}
	})
	mux.HandleFunc(admin.UrlNodeDisconnect, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node disconnect")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jerr.Get("error reading node disconnect body", err).Print()
			return
		}
		var disconnectRequest = new(admin.NodeDisconnectRequest)
		if err := json.Unmarshal(body, disconnectRequest); err != nil {
			jerr.Get("error unmarshalling node disconnect request", err).Print()
			return
		}
		for id, serverNode := range s.Nodes.Nodes {
			if id == disconnectRequest.NodeId {
				serverNode.Disconnect()
				fmt.Fprint(w, "Server disconnected")
				return
			}
		}
	})
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
		Port: config.GetAdminPort(),
	}
}
