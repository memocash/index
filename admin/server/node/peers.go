package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/config"
)

var peersRoute = admin.Route{
	Pattern: admin.UrlNodePeers,
	Handler: func(r admin.Response) {
		jlog.Logf("Beginning of processing node peers...\n")
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
		var peersResponse = &admin.NodePeersResponse{
			Peers: foundPeers,
		}
		if err := json.NewEncoder(r.Writer).Encode(peersResponse); err != nil {
			jerr.Get("error writing json peers response data", err).Print()
			return
		}
	},
}
