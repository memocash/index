package node

import (
	"github.com/memocash/server/admin/admin"
)

var getAddrsRoute = admin.Route{
	Pattern: admin.UrlNodeGetAddrs,
	Handler: func(r admin.Response) {
		for _, serverNode := range r.NodeGroup.Nodes {
			serverNode.GetAddr()
		}
	},
}
