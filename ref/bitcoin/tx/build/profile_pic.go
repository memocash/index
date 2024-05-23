package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type ProfilePicRequest struct {
	Wallet Wallet
	Url    string
}

func ProfilePic(request ProfilePicRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.ProfilePic{
			Url: request.Url,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("error building set profile pic tx; %w", err)
	}
	return txs, nil
}
