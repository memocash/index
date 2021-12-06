package build

import (
	"github.com/jchavannes/jgo/jerr"
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
		return nil, jerr.Get("error building user follow tx", err)
	}
	return txs, nil
}
