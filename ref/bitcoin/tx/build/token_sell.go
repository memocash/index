package build

import (
	"fmt"
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
		return nil, fmt.Errorf("error building token sell tx; %w", err)
	}
	return tx, nil
}
