package maint

import (
	"fmt"
	"log"
	"time"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type DoubleSpends struct {
	TotalEntries int
	DoubleSpends int
}

func (d *DoubleSpends) Check() error {
	lastStatus := time.Now()
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		var prevPrefix [36]byte
		for {
			if err := dbClient.GetWOpts(client.Opts{
				Topic: db.TopicChainOutputInput,
				Start: startUid,
				Max:   client.HugeLimit,
			}); err != nil {
				return fmt.Errorf("error getting output inputs for shard %d; %w", i, err)
			}
			d.TotalEntries += len(dbClient.Messages)
			for _, msg := range dbClient.Messages {
				var prefix [36]byte
				copy(prefix[:], msg.Uid[:36])
				if prefix == prevPrefix {
					/*var outputInput = new(chain.OutputInput)
					db.Set(outputInput, msg)
					log.Printf("Found double spend: %s:%d (spending: %s:%d)\n",
						chainhash.Hash(outputInput.Hash), outputInput.Index,
						chainhash.Hash(outputInput.PrevHash), outputInput.PrevIndex)*/
					d.DoubleSpends++
				}
				prevPrefix = prefix
			}
			if time.Since(lastStatus) >= 20*time.Second {
				log.Printf("Shard %d progress: %d entries, %d double spends so far\n", i, d.TotalEntries, d.DoubleSpends)
				lastStatus = time.Now()
			}
			if len(dbClient.Messages) < client.HugeLimit {
				break
			}
			startUid = dbClient.Messages[len(dbClient.Messages)-1].Uid
		}
		log.Printf("Shard %d: %d entries, %d double spends\n", i, d.TotalEntries, d.DoubleSpends)
	}
	return nil
}
