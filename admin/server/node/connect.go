package node

import (
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
	"io/ioutil"
)

var connectRoute = admin.Route{
	Pattern: admin.UrlNodeConnect,
	Handler: func(r admin.Response) {
		jlog.Log("Node connect")
		body, err := ioutil.ReadAll(r.Request.Body)
		if err != nil {
			jerr.Get("error reading node connect body", err).Print()
			return
		}
		var connectRequest = new(admin.NodeConnectRequest)
		if err := json.Unmarshal(body, connectRequest); err != nil {
			jerr.Get("error unmarshalling node connect request", err).Print()
			return
		}
		r.NodeGroup.AddNode(connectRequest.Ip, connectRequest.Port)
	},
}

var connectDefaultRoute = admin.Route{
	Pattern: admin.UrlNodeConnectDefault,
	Handler: func(r admin.Response) {
		jlog.Log("Node connect default")
		r.NodeGroup.AddDefaultNode()
	},
}

var connectNextRoute = admin.Route{
	Pattern: admin.UrlNodeConnectNext,
	Handler: func(r admin.Response) {
		jlog.Log("Node connect next")
		r.NodeGroup.AddNextNode()
	},
}

var disconnectRoute = admin.Route{
	Pattern: admin.UrlNodeDisconnect,
	Handler: func(r admin.Response) {
		jlog.Log("Node disconnect")
		body, err := ioutil.ReadAll(r.Request.Body)
		if err != nil {
			jerr.Get("error reading node disconnect body", err).Print()
			return
		}
		var disconnectRequest = new(admin.NodeDisconnectRequest)
		if err := json.Unmarshal(body, disconnectRequest); err != nil {
			jerr.Get("error unmarshalling node disconnect request", err).Print()
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
