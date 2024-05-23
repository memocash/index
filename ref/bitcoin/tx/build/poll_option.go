package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type PollOptionRequest struct {
	Wallet     Wallet
	PollTxHash []byte
	Option     string
}

func PollOption(request PollOptionRequest) (*memo.Tx, error) {
	tx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.PollOption{
			PollTxHash: request.PollTxHash,
			Option:     request.Option,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("error building poll create tx; %w", err)
	}
	return tx, nil
}
