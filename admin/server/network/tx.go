package network

import (
	"github.com/memocash/server/admin/admin"
)

var txRoute = admin.Route{
	Pattern: admin.UrlNetworkTx,
	Handler: func(r admin.Response) {
		/*var request = new(admin.NetworkTxRequest)
		if err := request.Parse(r.Request.Body); err != nil {
			r.Error(jerr.Get("error unmarshalling network tx request", err))
			return
		}
		txBlocks, err := item.GetSingleTxBlocks(request.HashByte)
		if err != nil {
			r.Error(jerr.Get("error getting tx blocks for hash", err))
			return
		}*/
	},
}
