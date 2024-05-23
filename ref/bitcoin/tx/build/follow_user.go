package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type FollowUserRequest struct {
	Wallet     Wallet
	UserPkHash []byte
	Unfollow   bool
}

func FollowUser(request FollowUserRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.FollowUser{
			UserPkHash: request.UserPkHash,
			Unfollow:   request.Unfollow,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("error building user follow tx; %w", err)
	}
	return txs, nil
}
