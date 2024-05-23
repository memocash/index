package node

import (
	"encoding/json"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
	"log"
)

var foundPeersRoute = admin.Route{
	Pattern: admin.UrlNodeFoundPeers,
	Handler: func(r admin.Response) {
		var request = new(admin.NodeFoundPeersRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(request); err != nil {
			log.Printf("error unmarshalling found peers request; %v", err)
			return
		}
		var foundFoundPeers []*item.FoundPeer
		var startId []byte
		var shard uint32
	FoundPeersLoop:
		for {
			foundPeers, err := item.GetFoundPeers(shard, startId, request.Ip, request.Port)
			if err != nil {
				log.Fatalf("fatal error getting found peers; %v", err)
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
			log.Printf("error marshalling and writing found peers response data; %v", err)
			return
		}
	},
}
