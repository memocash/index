package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type TopicFollowRequest struct {
	Wallet    Wallet
	TopicName string
	Unfollow  bool
}

func TopicFollow(request TopicFollowRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.TopicFollow{
			TopicName: request.TopicName,
			Unfollow:  request.Unfollow,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("error building topic follow tx; %w", err)
	}
	return txs, nil
}
