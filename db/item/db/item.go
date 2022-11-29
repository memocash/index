package db

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

func GetItem(obj Object) error {
	shard := obj.GetShard()
	shardConfig := config.GetShardConfig(uint32(shard), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(obj.GetTopic(), obj.GetUid()); err != nil && !client.IsMessageNotSetError(err) {
		return jerr.Get("error getting db item single", err)
	}
	if len(dbClient.Messages) != 1 {
		return jerr.Get("error item not found", client.EntryNotFoundError)
	}
	obj.Deserialize(dbClient.Messages[0].Message)
	return nil
}
