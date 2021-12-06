package node

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"net"
	"time"
)

type Group struct {
	Nodes      map[string]*Server
	Looping    bool
	LastPeerId []byte
	StartTime  time.Time
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
	var peerToUse *item.Peer
	for attempt := 1; ; attempt++ {
		newPeer, err := item.GetNextPeer(0, g.LastPeerId)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return jerr.Get("error getting next peer", err)
		}
		if newPeer == nil {
			return jerr.Newf("error unable to find usable new peer, %d attempts", attempt)
		}
		jlog.Logf("newPeer: %x\n", newPeer.GetUid())
		jlog.Logf("newPeer: %s:%d\n", net.IP(newPeer.Ip), newPeer.Port)
		g.LastPeerId = newPeer.GetUid()
		peerConnection, err := item.GetPeerConnectionLast(newPeer.Ip, newPeer.Port)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return jerr.Get("error getting last peer connection for new peer", err)
		}
		if peerConnection != nil {
			jlog.Logf("peerConnection: %s:%d - %s %s\n", net.IP(peerConnection.Ip), peerConnection.Port,
				peerConnection.Time.Format("2006-01-02 15:04:05"), peerConnection.Status)
		} else {
			jlog.Log("no peer connection found")
		}
		if peerConnection == nil || peerConnection.Time.Before(g.StartTime) {
			peerToUse = newPeer
			jlog.Logf("Found new peer after %d attempts\n", attempt)
			break
		}
	}
	jlog.Logf("peerToUse: %s:%d\n", net.IP(peerToUse.Ip), peerToUse.Port)
	/*if len(g.Nodes) > 1000 {
		jerr.Newf("fatal exiting").Fatal()
	}*/
	if peerToUse == nil {
		return jerr.New("error no peer found")
	}
	g.AddNode(peerToUse.Ip, peerToUse.Port)
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
