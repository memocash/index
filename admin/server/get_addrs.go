package server

import (
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
)

var getAddrsRoute = admin.Route{
	Pattern: admin.UrlNodeGetAddrs,
	Handler: func(r admin.Response) {
		jlog.Log("Node get addrs request")
		for _, serverNode := range r.NodeGroup.Nodes {
			serverNode.GetAddr()
		}
	},
}
