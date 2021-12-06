package build

import (
	"github.com/jchavannes/jgo/jerr"
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
		return nil, jerr.Get("error building mute user tx", err)
	}
	return txs, nil
}
