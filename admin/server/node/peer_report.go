package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/admin"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
)

var peerReportRoute = admin.Route{
	Pattern: admin.UrlNodePeerReport,
	Handler: func(r admin.Response) {
		var response = new(admin.NodePeerReportResponse)
		countPeers, err := item.GetCountPeers()
		if err != nil {
			r.Error(fmt.Errorf("error getting count peers; %w", err))
			return
		}
		countPeerConnections, err := item.GetCountPeerConnections()
		if err != nil {
			r.Error(fmt.Errorf("error getting count peer connections; %w", err))
			return
		}
		type PeerInfo struct {
			IpPort      []byte
			Connections uint
			Failed      uint
			Success     uint
		}
		var AddStats = func(i *PeerInfo) {
			if i.Connections == 0 {
				return
			}
			response.PeersAttempted++
			if i.Success > 0 {
				response.PeersConnected++
			} else {
				response.PeersFailed++
			}
		}
		var peerInfo = new(PeerInfo)
		for shard := uint32(0); shard < config.GetTotalShards(); shard++ {
			for startId := []byte{}; ; {
				peerConnections, err := item.GetPeerConnections(item.PeerConnectionsRequest{
					Shard:   shard,
					StartId: startId,
					Max:     client.LargeLimit,
				})
				if err != nil {
					r.Error(fmt.Errorf("fatal error getting peer connections; %w", err))
					return
				}
				for i, peerConnection := range peerConnections {
					if i == 0 && bytes.Equal(peerConnection.GetUid(), startId) {
						continue
					}
					ipPort := jutil.CombineBytes(peerConnection.Ip, jutil.GetUintData(uint(peerConnection.Port)))
					if !bytes.Equal(peerInfo.IpPort, ipPort) {
						AddStats(peerInfo)
						peerInfo = new(PeerInfo)
						peerInfo.IpPort = ipPort
					}
					peerInfo.Connections++
					if peerConnection.Status == item.PeerConnectionStatusSuccess {
						peerInfo.Success++
					} else {
						peerInfo.Failed++
					}
				}
				if len(peerConnections) < client.LargeLimit {
					break
				}
				startId = peerConnections[len(peerConnections)-1].GetUid()
			}
		}
		AddStats(peerInfo)
		response.TotalPeers = countPeers
		response.TotalAttempts = countPeerConnections
		if err := json.NewEncoder(r.Writer).Encode(response); err != nil {
			r.Error(fmt.Errorf("error marshalling and writing peer report response data; %w", err))
			return
		}
	},
}
