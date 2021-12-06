package node

import (
	"fmt"
	"github.com/memocash/index/admin/admin"
	"net"
)

var listConnectionsRoute = admin.Route{
	Pattern: admin.UrlNodeListConnections,
	Handler: func(r admin.Response) {
		for id, serverNode := range r.NodeGroup.Nodes {
			fmt.Fprintf(r.Writer, "%s - %s:%d (%t)\n", id, net.IP(serverNode.Ip), serverNode.Port,
				serverNode.Peer != nil && serverNode.Peer.Connected())
		}
	},
}
