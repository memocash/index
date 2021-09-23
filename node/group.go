package node

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
)

type Group struct {
	Nodes map[string]*Server
}

func (g *Group) AddDefaultNode() {
	g.AddNode(GetLocalhost(), DefaultPort)
}

func (g *Group) AddNode(ip []byte, port uint16) {
	nodeId := groupNodeId(ip, port)
	if _, exists := g.Nodes[nodeId]; exists {
		return
	}
	g.Nodes[nodeId] = NewServer(ip, port)
	go func() {
		err := g.Nodes[nodeId].Run()
		if err != nil {
			jerr.Get("error node failed", err).Print()
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
