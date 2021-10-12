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

var foundPeersRoute = admin.Route{
	Pattern: admin.UrlNodeFoundPeers,
	Handler: func(r admin.Response) {
		jlog.Log("Node found peers request")
		body, err := ioutil.ReadAll(r.Request.Body)
		if err != nil {
			jerr.Get("error reading node found peers request", err).Print()
			return
		}
		var request = new(admin.NodeFoundPeersRequest)
		if err := json.Unmarshal(body, request); err != nil {
			jerr.Get("error unmarshalling found peers request", err).Print()
			return
		}
		var foundFoundPeers []*item.FoundPeer
		var startId []byte
		var shard uint32
	FoundPeersLoop:
		for {
			foundPeers, err := item.GetFoundPeers(shard, startId, request.Ip, request.Port)
			if err != nil {
				jerr.Get("fatal error getting found peers", err).Fatal()
			}
			for _, foundPeer := range foundPeers {
				foundFoundPeers = append(foundFoundPeers, foundPeer)
				if len(foundFoundPeers) >= client.LargeLimit {
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
		var foundPeersResponse = &admin.NodeFoundPeersResponse{
			FoundPeers: foundFoundPeers,
		}
		foundPeersResponseData, err := json.Marshal(foundPeersResponse)
		if err != nil {
			jerr.Get("error marshalling found peers response data", err).Print()
			return
		}
		if _, err = r.Writer.Write(foundPeersResponseData); err != nil {
			jerr.Get("error writing found peers response data", err).Print()
			return
		}
	},
}
