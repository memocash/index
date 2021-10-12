package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/config"
	"io/ioutil"
)

var historyRoute = admin.Route{
	Pattern: admin.UrlNodeHistory,
	Handler: func(r admin.Response) {
		jlog.Log("Node list history")
		body, err := ioutil.ReadAll(r.Request.Body)
		if err != nil {
			jerr.Get("error reading node history body", err).Print()
			return
		}
		var historyRequest = new(admin.NodeHistoryRequest)
		if err := json.Unmarshal(body, historyRequest); err != nil {
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
		historyResponseData, err := json.Marshal(historyResponse)
		if err != nil {
			jerr.Get("error marshalling history response data", err).Print()
			return
		}
		if _, err = r.Writer.Write(historyResponseData); err != nil {
			jerr.Get("error writing history response data", err).Print()
			return
		}
	},
}
