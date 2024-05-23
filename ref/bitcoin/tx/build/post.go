package build

import (
	"fmt"
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
		return nil, fmt.Errorf("error building post tx; %w", err)
	}
	return txs, nil
}
