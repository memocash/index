package node

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/item"
)

type Group struct {
	Nodes      map[string]*Server
	Looping    bool
	LastPeerId []byte
}

func (g Group) HasActive() bool {
	for _, node := range g.Nodes {
		if node.Peer != nil && node.Peer.Connected() {
			return true
		}
	}
	return false
}

func (g *Group) AddDefaultNode() {
	g.AddNode(GetLocalhost(), DefaultPort)
}

func (g *Group) AddNextNode() error {
	newPeer, err := item.GetNextPeer(0, g.LastPeerId)
	if err != nil {
		return jerr.Get("error getting next peer", err)
	}
	g.LastPeerId = newPeer.GetUid()
	g.AddNode(newPeer.Ip, newPeer.Port)
	return nil
}

func (g *Group) AddNode(ip []byte, port uint16) {
	nodeId := groupNodeId(ip, port)
	if _, exists := g.Nodes[nodeId]; exists {
		if g.Nodes[nodeId].Peer != nil && g.Nodes[nodeId].Peer.Connected() {
			return
		}
	}
	g.Nodes[nodeId] = NewServer(ip, port)
	go func() {
		if err := g.Nodes[nodeId].Run(); err != nil {
			jerr.Get("error node failed", err).Print()
		}
		if g.Looping && !g.HasActive() {
			if err := g.AddNextNode(); err != nil {
				jerr.Get("error adding next node in looper", err).Print()
			}
		}
	}()
}

func NewGroup() *Group {
	return &Group{
		Nodes: make(map[string]*Server),
	}
}

func groupNodeId(ip []byte, port uint16) string {
	return fmt.Sprintf("%x-%d", ip, port)
}
