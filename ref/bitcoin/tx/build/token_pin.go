package build

import (
	"fmt"
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
		return nil, fmt.Errorf("error building token pin tx; %w", err)
	}
	return pinTx, nil
}
