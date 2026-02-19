package db

import (
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

func GetShardClient(shard uint32) *client.Client {
	return client.NewClient(config.GetShardConfig(shard, config.GetQueueShards()).GetHost())
}
