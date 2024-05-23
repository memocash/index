package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
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
		return nil, fmt.Errorf("error building set profile tx; %w", err)
	}
	return txs, nil
}
