package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client/peer"
	"net"
)

var peersRoute = admin.Route{
	Pattern: admin.UrlNodePeers,
	Handler: func(r admin.Response) {
		var request = new(admin.NodePeersRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(request); err != nil {
			r.Error(jerr.Get("error unmarshalling peers request", err))
			return
		}
		peerList := peer.NewList()
		if err := peerList.GetPeers(request.Filter); err != nil {
			r.Error(jerr.Get("error getting list of peers with filter", err))
			return
		}
		var responsePeers = make([]*admin.Peer, len(peerList.Peers))
		for i := range peerList.Peers {
			responsePeers[i] = &admin.Peer{
				Ip:       net.IP(peerList.Peers[i].Ip).String(),
				Port:     peerList.Peers[i].Port,
				Services: peerList.Peers[i].Services,
				Time:     peerList.Peers[i].Time,
				Status:   peerList.Peers[i].Status,
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
