package server

import (
	"encoding/json"
	"github.com/memocash/index/admin/admin"
	"log"
)

var indexRoute = admin.Route{
	Pattern: admin.UrlIndex,
	Handler: func(r admin.Response) {
		if err := json.NewEncoder(r.Writer).Encode(struct {
			Name    string
			Version string
		}{
			Name:    "Memo Admin",
			Version: "0.1",
		}); err != nil {
			log.Printf("error marshalling and writing memo admin version; %v", err)
			return
		}
	},
}
