package server

import (
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
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
			jerr.Get("error marshalling and writing memo admin version", err).Print()
			return
		}
	},
}
