package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
	"net"
)

var historyRoute = admin.Route{
	Pattern: admin.UrlNodeHistory,
	Handler: func(r admin.Response) {
		var historyRequest = new(admin.NodeHistoryRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(historyRequest); err != nil {
			jerr.Get("error unmarshalling node history request", err).Print()
			return
		}
		var foundPeerConnections []*item.PeerConnection
		var startId []byte
		var shard uint32
	PeerConnectionsLoop:
		for {
			peerConnections, err := item.GetPeerConnections(item.PeerConnectionsRequest{
				Shard:   shard,
				StartId: startId,
				Ip:      net.ParseIP(historyRequest.Ip),
				Port:    historyRequest.Port,
			})
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
		var connections = make([]admin.Connection, len(foundPeerConnections))
		for i := range foundPeerConnections {
			connections[i] = admin.Connection{
				Ip:     net.IP(foundPeerConnections[i].Ip).String(),
				Port:   foundPeerConnections[i].Port,
				Time:   foundPeerConnections[i].Time,
				Status: foundPeerConnections[i].Status,
			}
		}
		var historyResponse = &admin.NodeHistoryResponse{
			Connections: connections,
		}
		if err := json.NewEncoder(r.Writer).Encode(historyResponse); err != nil {
			jerr.Get("error marshalling and writing history response data", err).Print()
			return
		}
	},
}
