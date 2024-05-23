package network

import (
	"github.com/memocash/index/admin/admin"
)

var txRoute = admin.Route{
	Pattern: admin.UrlNetworkTx,
	Handler: func(r admin.Response) {
		/*var request = new(admin.NetworkTxRequest)
		if err := request.Parse(r.Request.Body); err != nil {
			r.Error(fmt.Errorf("error unmarshalling network tx request; %w", err))
			return
		}
		txBlocks, err := item.GetSingleTxBlocks(request.HashByte)
		if err != nil {
			r.Error(fmt.Errorf("error getting tx blocks for hash; %w", err))
			return
		}*/
	},
}
