package admin

import (
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/node"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
	"net"
	"net/http"
)

type NodeDisconnectRequest struct {
	NodeId string
}

const (
	UrlIndex               = "/"
	UrlNodeGetAddrs        = "/node/get_addrs"
	UrlNodeConnectDefault  = "/node/connect_default"
	UrlNodeListConnections = "/node/list_connections"
	UrlNodeDisconnect      = "/node/disconnect"
)

type Server struct {
	Nodes *node.Group
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(UrlIndex, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Memo Admin 0.1")
	})
	mux.HandleFunc(UrlNodeGetAddrs, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node get addrs request")
		for _, serverNode := range s.Nodes.Nodes {
			serverNode.GetAddr()
		}
	})
	mux.HandleFunc(UrlNodeConnectDefault, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node connect default")
		s.Nodes.AddDefaultNode()
	})
	mux.HandleFunc(UrlNodeListConnections, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node list connections")
		for id, serverNode := range s.Nodes.Nodes {
			fmt.Fprintf(w, "%s - %s:%d\n", id, net.IP(serverNode.Ip), serverNode.Port)
		}
	})
	mux.HandleFunc(UrlNodeDisconnect, func(w http.ResponseWriter, r *http.Request) {
		jlog.Log("Node disconnect")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jerr.Get("error reading node disconnect body", err).Print()
			return
		}
		var disconnectRequest = new(NodeDisconnectRequest)
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
		Addr:    config.GetHost(config.GetAdminPort()),
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
	}
}
