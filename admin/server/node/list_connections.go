package node

import (
	"fmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
	"net"
)

var listConnectionsRoute = admin.Route{
	Pattern: admin.UrlNodeListConnections,
	Handler: func(r admin.Response) {
		jlog.Log("Node list connections")
		for id, serverNode := range r.NodeGroup.Nodes {
			fmt.Fprintf(r.Writer, "%s - %s:%d (%t)\n", id, net.IP(serverNode.Ip), serverNode.Port,
				serverNode.Peer != nil && serverNode.Peer.Connected())
		}
	},
}
