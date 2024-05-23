package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type TopicMessageRequest struct {
	Wallet    Wallet
	TopicName string
	Message   string
}

func TopicMessage(request TopicMessageRequest) ([]*memo.Tx, error) {
	txs, err := Simple(request.Wallet, []*memo.Output{{
		Script: &script.TopicMessage{
			TopicName: request.TopicName,
			Message:   request.Message,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("error building topic message tx; %w", err)
	}
	return txs, nil
}
