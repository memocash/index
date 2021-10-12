package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/config"
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
		if err := json.NewEncoder(r.Writer).Encode(historyResponse); err != nil {
			jerr.Get("error marshalling and writing history response data", err).Print()
			return
		}
	},
}
