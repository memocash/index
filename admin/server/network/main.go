package network

import "github.com/memocash/server/admin/admin"

func GetRoutes() []admin.Route {
	return []admin.Route{
		txRoute,
	}
}
