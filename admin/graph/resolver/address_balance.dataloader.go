package resolver

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/node/obj/get"
	"time"
)

var addressBalanceLoaderConfig = dataloader.AddressBalanceLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(addressStrings []string) ([]int64, []error) {
		var balances = make([]int64, len(addressStrings))
		var errors = make([]error, len(addressStrings))
		for i := range addressStrings {
			balance, err := get.NewBalanceFromAddress(addressStrings[i])
			if err != nil {
				errors[i] = jerr.Get("error getting address from string for balance", err)
				continue
			}
			if err := balance.GetBalanceByUtxos(); err != nil {
				errors[i] = jerr.Get("error getting address balance from network", err)
				continue
			}
			balances[i] = balance.Balance
		}
		return balances, errors
	},
}
