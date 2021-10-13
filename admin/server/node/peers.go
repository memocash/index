package node

import (
	"bytes"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/config"
	"net"
)

var peersRoute = admin.Route{
	Pattern: admin.UrlNodePeers,
	Handler: func(r admin.Response) {
		var request = new(admin.NodePeersRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(request); err != nil {
			jerr.Get("error unmarshalling peers request", err).Print()
			return
		}
		var foundPeers []*item.Peer
		var startId []byte
		var shard uint32
	FoundPeersLoop:
		for {
			peers, err := item.GetPeers(shard, startId)
			if err != nil {
				jerr.Get("fatal error getting peers", err).Fatal()
			}
			for _, peer := range peers {
				foundPeers = append(foundPeers, peer)
				if len(foundPeers) >= client.LargeLimit {
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
		var ipPorts = make([]item.IpPort, len(foundPeers))
		for i := range foundPeers {
			ipPorts[i] = item.IpPort{
				Ip:   foundPeers[i].Ip,
				Port: foundPeers[i].Port,
			}
		}
		peerConnectionLasts, err := item.GetPeerConnectionLasts(ipPorts)
		if err != nil {
			jerr.Get("error getting peer connection lasts for found peers", err).Print()
			return
		}
		var responsePeers = make([]*admin.Peer, len(foundPeers))
		for i := range foundPeers {
			responsePeers[i] = &admin.Peer{
				Ip:       net.IP(foundPeers[i].Ip).String(),
				Port:     foundPeers[i].Port,
				Services: foundPeers[i].Services,
			}
			for _, peerConnectionLast := range peerConnectionLasts {
				if bytes.Equal(peerConnectionLast.Ip, foundPeers[i].Ip) && peerConnectionLast.Port == foundPeers[i].Port {
					responsePeers[i].Time = peerConnectionLast.Time
					responsePeers[i].Status = peerConnectionLast.Status
					break
				}
			}
		}
		var peersResponse = &admin.NodePeersResponse{
			Peers: responsePeers,
		}
		if err := json.NewEncoder(r.Writer).Encode(peersResponse); err != nil {
			jerr.Get("error writing json peers response data", err).Print()
			return
		}
	},
}
