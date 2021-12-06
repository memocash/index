package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type TokenSellRequest struct {
	Wallet Wallet
	InOuts []script.InOut
}

func TokenSell(request TokenSellRequest) (*memo.Tx, error) {
	tx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.TokenSell{
			InOuts: request.InOuts,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building token sell tx", err)
	}
	return tx, nil
}
