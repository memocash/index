package peer

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/config"
	"time"
)

const (
	FilterAll       = "all"
	FilterAttempted = "attempted"
	FilterSuccesses = "successes"
)

type Peer struct {
	Ip       []byte
	Port     uint16
	Services uint64
	Time     time.Time
	Status   item.PeerConnectionStatus
}

type List struct {
	Peers []*Peer
}

func (l *List) GetPeers(filter string) error {
	var startId []byte
	var shard uint32
	for {
		peers, err := item.GetPeers(shard, startId)
		if err != nil {
			return jerr.Get("error getting next set of peers", err)
		}
		var ipPorts = make([]item.IpPort, len(peers))
		for i := range peers {
			ipPorts[i] = item.IpPort{
				Ip:   peers[i].Ip,
				Port: peers[i].Port,
			}
		}
		peerConnectionLasts, err := item.GetPeerConnectionLasts(ipPorts)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return jerr.Get("error getting peer connection lasts for found peers", err)
		}
		for _, peer := range peers {
			var newPeer = &Peer{
				Ip:       peer.Ip,
				Port:     peer.Port,
				Services: peer.Services,
			}
			for i, peerConnectionLast := range peerConnectionLasts {
				if bytes.Equal(peer.Ip, peerConnectionLast.Ip) && peer.Port == peerConnectionLast.Port {
					newPeer.Time = peerConnectionLast.Time
					newPeer.Status = peerConnectionLast.Status
					peerConnectionLasts = append(peerConnectionLasts[:i], peerConnectionLasts[i+1:]...)
					break
				}
			}
			switch filter {
			case FilterAttempted:
				if newPeer.Time.IsZero() {
					continue
				}
			case FilterSuccesses:
				if newPeer.Status != item.PeerConnectionStatusSuccess {
					continue
				}
			case FilterAll:
			}
			l.Peers = append(l.Peers, newPeer)
			if len(l.Peers) >= client.LargeLimit {
				break
			}
		}
		if len(peers) < client.LargeLimit {
			shard++
			if shard >= config.GetTotalShards() {
				break
			}
			startId = nil
		} else {
			startId = peers[len(peers)-1].GetUid()
		}
	}
	return nil
}

func NewList() *List {
	return &List{}
}
