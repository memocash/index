package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
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
		return nil, jerr.Get("error building topic follow tx", err)
	}
	return txs, nil
}
