package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

type ReplyRequest struct {
	Wallet  Wallet
	TxHash  []byte
	Message string
}

func Reply(request ReplyRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.Reply{
			TxHash:  request.TxHash,
			Message: request.Message,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building reply tx", err)
	}
	return txs, nil
}
