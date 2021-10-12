package server

import (
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/admin"
)

var indexRoute = admin.Route{
	Pattern: admin.UrlIndex,
	Handler: func(r admin.Response) {
		jsonData, err := json.Marshal(struct {
			Name    string
			Version string
		}{
			Name:    "Memo Admin",
			Version: "0.1",
		})
		if err != nil {
			jerr.Get("error marshalling memo admin version", err).Print()
			return
		}
		fmt.Fprint(r.Writer, string(jsonData))
	},
}
