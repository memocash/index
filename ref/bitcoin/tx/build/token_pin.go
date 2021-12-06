package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type TokenPinRequest struct {
	Wallet      Wallet
	PostTxHash  []byte
	SendTxHash  []byte
	SendTxIndex uint
}

func TokenPin(request TokenPinRequest) (*memo.Tx, error) {
	pinTx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.TokenPin{
			PostTxHash:  request.PostTxHash,
			TokenTxHash: request.SendTxHash,
			TokenIndex:  request.SendTxIndex,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building token pin tx", err)
	}
	return pinTx, nil
}
