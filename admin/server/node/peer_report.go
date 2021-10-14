package node

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
)

var peerReportRoute = admin.Route{
	Pattern: admin.UrlNodePeerReport,
	Handler: func(r admin.Response) {
		var response = new(admin.NodePeerReportResponse)
		if err := json.NewEncoder(r.Writer).Encode(response); err != nil {
			jerr.Get("error marshalling and writing peer report response data", err).Print()
			return
		}
	},
}
