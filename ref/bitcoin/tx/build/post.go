package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type PostRequest struct {
	Wallet  Wallet
	Message string
}

func Post(request PostRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.Post{
			Message: request.Message,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building post tx", err)
	}
	return txs, nil
}
