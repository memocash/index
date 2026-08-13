package db

import (
	"context"
	"fmt"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

// GetFiltered runs one pattern-filtered call against a single shard,
// returning up to limit matching messages. Callers page keyset-style: pass
// the last returned uid + 0x00 as pattern.Start on the next call and stop on
// a short page. A sparse match can make a single call scan a long range
// server-side, so pass client.NewOptionTimeout when the topic is large. A
// fresh client is used so concurrent shard scans don't share reply state.
func GetFiltered(ctx context.Context, topic string, shard uint32, pattern client.FilterPattern, limit int,
	opts ...client.Option) ([]client.Message, error) {
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if limit > 0 {
		opts = append(opts, client.NewOptionLimit(limit))
	}
	if err := dbClient.GetFiltered(ctx, topic, []client.FilterPattern{pattern}, opts...); err != nil {
		return nil, fmt.Errorf("error getting filtered messages shard %d; %w", shard, err)
	}
	return dbClient.Messages, nil
}
