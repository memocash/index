package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type AliasRequest struct {
	Wallet  Wallet
	Address wallet.Address
	Alias   string
}

func Alias(request AliasRequest) (*memo.Tx, error) {
	tx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.Alias{
			Hash:  request.Address.GetPkHash(),
			Alias: request.Alias,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building alias tx", err)
	}
	return tx, nil
}
