package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

type ProfileRequest struct {
	Wallet Wallet
	Text   string
}

func Profile(request ProfileRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.Profile{
			Text: request.Text,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building set profile tx", err)
	}
	return txs, nil
}
