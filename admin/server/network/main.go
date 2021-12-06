package network

import "github.com/memocash/index/admin/admin"

func GetRoutes() []admin.Route {
	return []admin.Route{
		txRoute,
	}
}
