package queue

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

type Get struct {
	Shard uint32
	Items []Item
}

func (r *Get) GetByPrefixes(ctx context.Context, topic string, prefixes [][]byte) error {
	shardConfig := config.GetShardConfig(r.Shard, config.GetQueueShards())
	db := client.NewClient(fmt.Sprintf("127.0.0.1:%d", shardConfig.Port))
	var clientPrefixes = make([]client.Prefix, len(prefixes))
	for i := range prefixes {
		clientPrefixes[i] = client.NewPrefix(prefixes[i])
	}
	if err := db.GetByPrefixes(ctx, topic, clientPrefixes); err != nil {
		return fmt.Errorf("error getting by prefixes using queue client; %w", err)
	}
	r.Items = make([]Item, len(db.Messages))
	for i := range db.Messages {
		r.Items[i] = Item{
			Topic: db.Messages[i].Topic,
			Uid:   db.Messages[i].Uid,
			Data:  db.Messages[i].Message,
		}
	}
	return nil
}

func NewGet(shard uint32) *Get {
	return &Get{
		Shard: shard,
	}
}
