package node

import (
	"github.com/memocash/server/admin/admin"
)

var loopingEnableRoute = admin.Route{
	Pattern: admin.UrlNodeLoopingEnable,
	Handler: func(r admin.Response) {
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
		r.NodeGroup.Looping = false
	},
}
