package node

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"log"
	"net"
	"time"
)

type Group struct {
	Context    context.Context
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
		newPeer, err := item.GetNextPeer(g.Context, 0, g.LastPeerId)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return fmt.Errorf("error getting next peer; %w", err)
		}
		if newPeer == nil {
			return fmt.Errorf("error unable to find usable new peer, %d attempts", attempt)
		}
		log.Printf("newPeer: %x\n", newPeer.GetUid())
		log.Printf("newPeer: %s:%d\n", net.IP(newPeer.Ip), newPeer.Port)
		g.LastPeerId = newPeer.GetUid()
		peerConnection, err := item.GetPeerConnectionLast(g.Context, newPeer.Ip, newPeer.Port)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return fmt.Errorf("error getting last peer connection for new peer; %w", err)
		}
		if peerConnection != nil {
			log.Printf("peerConnection: %s:%d - %s %s\n", net.IP(peerConnection.Ip), peerConnection.Port,
				peerConnection.Time.Format("2006-01-02 15:04:05"), peerConnection.Status)
		} else {
			log.Println("no peer connection found")
		}
		if peerConnection == nil || peerConnection.Time.Before(g.StartTime) {
			peerToUse = newPeer
			log.Printf("Found new peer after %d attempts\n", attempt)
			break
		}
	}
	log.Printf("peerToUse: %s:%d\n", net.IP(peerToUse.Ip), peerToUse.Port)
	/*if len(g.Nodes) > 1000 {
		log.Fatalf("fatal exiting")
	}*/
	if peerToUse == nil {
		return fmt.Errorf("error no peer found")
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
			log.Printf("error node failed; %v", err)
		}
		if g.Looping && !g.HasActive() {
			if err := g.AddNextNode(); err != nil {
				log.Printf("error adding next node in looper; %v", err)
			}
		}
	}()
}

func NewGroup(ctx context.Context) *Group {
	return &Group{
		Context: ctx,
		Nodes:   make(map[string]*Server),
	}
}

func groupNodeId(ip []byte, port uint16) string {
	return fmt.Sprintf("%x-%d", ip, port)
}
