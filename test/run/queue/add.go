package queue

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
	"time"
)

type Add struct {
	Shard uint32
}

func (a *Add) Add(items []Item) error {
	shardConfig := config.GetShardConfig(a.Shard, config.GetQueueShards())
	db := client.NewClient(fmt.Sprintf("127.0.0.1:%d", shardConfig.Port))
	var messages = make([]*client.Message, len(items))
	for i := range items {
		messages[i] = &client.Message{
			Topic:   items[i].Topic,
			Uid:     items[i].Uid,
			Message: items[i].Data,
		}
	}
	if err := db.Save(messages, time.Now()); err != nil {
		return jerr.Get("error saving queue client messages", err)
	}
	return nil
}

func NewAdd(shard uint32) *Add {
	return &Add{
		Shard: shard,
	}
}
