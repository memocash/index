package node

import (
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/admin/admin"
)

var loopingEnableRoute = admin.Route{
	Pattern: admin.UrlNodeLoopingEnable,
	Handler: func(r admin.Response) {
		jlog.Log("Node looping enable")
		if r.NodeGroup.Looping {
			return
		}
		r.NodeGroup.Looping = true
		if !r.NodeGroup.HasActive() {
			r.NodeGroup.AddNextNode()
		}
	},
}

var loopingDisableRoute = admin.Route{
	Pattern: admin.UrlNodeLoopingDisable,
	Handler: func(r admin.Response) {
		jlog.Log("Node looping disabled")
		r.NodeGroup.Looping = false
	},
}
