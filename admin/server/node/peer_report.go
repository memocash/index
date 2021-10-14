package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
	"github.com/memocash/server/db/item"
)

var peerReportRoute = admin.Route{
	Pattern: admin.UrlNodePeerReport,
	Handler: func(r admin.Response) {
		var response = new(admin.NodePeerReportResponse)
		countPeers, err := item.GetCountPeers()
		if err != nil {
			r.Error(jerr.Get("error getting count peers", err))
			return
		}
		countPeerConnections, err := item.GetCountPeerConnections()
		if err != nil {
			r.Error(jerr.Get("error getting count peer connections", err))
			return
		}
		response.TotalPeers = countPeers
		response.TotalAttempts = countPeerConnections
		if err := json.NewEncoder(r.Writer).Encode(response); err != nil {
			r.Error(jerr.Get("error marshalling and writing peer report response data", err))
			return
		}
	},
}
