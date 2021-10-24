package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

type SetNameRequest struct {
	Wallet Wallet
	Name   string
}

func SetName(request SetNameRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.SetName{
			Name: request.Name,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building set name tx", err)
	}
	return txs, nil
}
