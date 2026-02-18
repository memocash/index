package maint

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type DoubleSpend struct {
	Spends []*chain.OutputInput
}

type RandomDoubleSpend struct {
	DoubleSpends []DoubleSpend
}

func (r *RandomDoubleSpend) Find(ctx context.Context) error {
	shards := config.GetQueueShards()
	shardIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(shards))))
	if err != nil {
		return fmt.Errorf("error generating random shard; %w", err)
	}
	shardConfig := shards[shardIdx.Int64()]
	var startUid = make([]byte, 72)
	if _, err := rand.Read(startUid); err != nil {
		return fmt.Errorf("error generating random start; %w", err)
	}
	dbClient := client.NewClient(shardConfig.GetHost())
	for {
		if err := dbClient.GetByPrefix(ctx, db.TopicChainOutputInput, client.Prefix{
			Start: startUid,
		}, client.OptionLargeLimit()); err != nil {
			return fmt.Errorf("error getting output inputs; %w", err)
		}
		var prevOutputInput *chain.OutputInput
		for _, msg := range dbClient.Messages {
			outputInput := new(chain.OutputInput)
			db.Set(outputInput, msg)
			if prevOutputInput != nil &&
				outputInput.PrevHash == prevOutputInput.PrevHash &&
				outputInput.PrevIndex == prevOutputInput.PrevIndex {
				r.DoubleSpends = append(r.DoubleSpends, DoubleSpend{
					Spends: []*chain.OutputInput{
						prevOutputInput,
						outputInput,
					},
				})
			}
			prevOutputInput = outputInput
		}
		if len(dbClient.Messages) < client.LargeLimit || len(r.DoubleSpends) > 0 {
			break
		}
		startUid = dbClient.Messages[len(dbClient.Messages)-1].Uid
	}
	if len(r.DoubleSpends) > 0 {
		return nil
	}
	return fmt.Errorf("no double spend found scanning from random start")
}
