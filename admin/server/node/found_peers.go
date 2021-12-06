package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
)

var foundPeersRoute = admin.Route{
	Pattern: admin.UrlNodeFoundPeers,
	Handler: func(r admin.Response) {
		var request = new(admin.NodeFoundPeersRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(request); err != nil {
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
		if err := json.NewEncoder(r.Writer).Encode(foundPeersResponse); err != nil {
			jerr.Get("error marshalling and writing found peers response data", err).Print()
			return
		}
	},
}
