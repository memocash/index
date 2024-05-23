package node

import (
	"encoding/json"
	"fmt"
	"github.com/memocash/index/admin/admin"
	"log"
)

var connectRoute = admin.Route{
	Pattern: admin.UrlNodeConnect,
	Handler: func(r admin.Response) {
		var connectRequest = new(admin.NodeConnectRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(connectRequest); err != nil {
			log.Printf("error unmarshalling node connect request; %v", err)
			return
		}
		r.NodeGroup.AddNode(connectRequest.Ip, connectRequest.Port)
	},
}

var connectDefaultRoute = admin.Route{
	Pattern: admin.UrlNodeConnectDefault,
	Handler: func(r admin.Response) {
		r.NodeGroup.AddDefaultNode()
	},
}

var connectNextRoute = admin.Route{
	Pattern: admin.UrlNodeConnectNext,
	Handler: func(r admin.Response) {
		r.NodeGroup.AddNextNode()
	},
}

var disconnectRoute = admin.Route{
	Pattern: admin.UrlNodeDisconnect,
	Handler: func(r admin.Response) {
		var disconnectRequest = new(admin.NodeDisconnectRequest)
		if err := json.NewDecoder(r.Request.Body).Decode(disconnectRequest); err != nil {
			log.Printf("error unmarshalling node disconnect request; %v", err)
			return
		}
		for id, serverNode := range r.NodeGroup.Nodes {
			if id == disconnectRequest.NodeId {
				serverNode.Disconnect()
				fmt.Fprint(r.Writer, "Server disconnected")
				return
			}
		}
	},
}
