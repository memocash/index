package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type MuteUserRequest struct {
	Wallet     Wallet
	MutePkHash []byte
	Unmute     bool
}

func MuteUser(request MuteUserRequest) ([]*memo.Tx, error) {
	outputs := []*memo.Output{{
		Script: &script.MuteUser{
			MutePkHash: request.MutePkHash,
			Unmute:     request.Unmute,
		},
	}}
	txs, err := Simple(request.Wallet, outputs)
	if err != nil {
		return nil, fmt.Errorf("error building mute user tx; %w", err)
	}
	return txs, nil
}
