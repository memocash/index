package node

import "github.com/memocash/server/admin/admin"

func GetRoutes() []admin.Route {
	return []admin.Route{
		connectRoute,
		connectDefaultRoute,
		connectNextRoute,
		disconnectRoute,
		loopingEnableRoute,
		loopingDisableRoute,
		listConnectionsRoute,
		historyRoute,
		foundPeersRoute,
	}
}
