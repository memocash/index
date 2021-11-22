package queue

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/config"
	"time"
)

type Get struct {
	Shard uint32
	Items []Item
}

func (r *Get) GetByPrefixes(topic string, prefixes [][]byte) error {
	shardConfig := config.GetShardConfig(r.Shard, config.GetQueueShards())
	db := client.NewClient(fmt.Sprintf("127.0.0.1:%d", shardConfig.Port))
	if err := db.GetWOpts(client.Opts{
		Topic:    topic,
		Prefixes: prefixes,
	}); err != nil {
		return jerr.Get("error getting by prefixes using queue client", err)
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

func (r *Get) GetAndWait(topic string, start []byte) error {
	shardConfig := config.GetShardConfig(r.Shard, config.GetQueueShards())
	db := client.NewClient(fmt.Sprintf("127.0.0.1:%d", shardConfig.Port))
	if err := db.GetWOpts(client.Opts{
		Topic:   topic,
		Start:   start,
		Wait:    true,
		Timeout: time.Second,
	}); err != nil {
		return jerr.Get("error getting and waiting with start using queue client", err)
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
