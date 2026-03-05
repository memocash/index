package maint

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type ListHeightDuplicates struct {
	Total int
}

func (l *ListHeightDuplicates) List(ctx context.Context) error {
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		for {
			if err := dbClient.GetByPrefix(ctx, db.TopicChainHeightDuplicate, client.Prefix{
				Start: startUid,
			}, client.OptionHugeLimit()); err != nil {
				return fmt.Errorf("error getting height duplicates for shard %d; %w", i, err)
			}
			for _, msg := range dbClient.Messages {
				var heightDuplicate chain.HeightDuplicate
				heightDuplicate.SetUid(msg.Uid)
				log.Printf("Height: %d, Block: %s\n", heightDuplicate.Height, chainhash.Hash(heightDuplicate.BlockHash).String())
				l.Total++
			}
			if len(dbClient.Messages) < client.HugeLimit {
				break
			}
			startUid = dbClient.Messages[len(dbClient.Messages)-1].Uid
		}
	}
	return nil
}
