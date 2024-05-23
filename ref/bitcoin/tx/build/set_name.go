package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
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
		return nil, fmt.Errorf("error building set name tx; %w", err)
	}
	return txs, nil
}
