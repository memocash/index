package block_tx

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
)

type LoopRaw struct {
	Processor func([]*item.BlockTxRaw) error
}

func (l *LoopRaw) Process(blockHash []byte) error {
	const limit = client.DefaultLimit
	for _, shard := range config.GetQueueShards() {
		var startTxHash []byte
		for {
			blockTxes, err := item.GetBlockTxesRaw(item.BlockTxesRawRequest{
				Shard:       shard.Min,
				BlockHash:   blockHash,
				StartTxHash: startTxHash,
				Limit:       limit,
			})
			if err != nil {
				return jerr.Get("error getting block txes raw for loop process", err)
			}
			if err := l.Processor(blockTxes); err != nil {
				return jerr.Get("error processing block txes raw", err)
			}
			if len(blockTxes) < limit {
				break
			}
			startTxHash = blockTxes[len(blockTxes)-1].TxHash
		}
	}
	return nil
}

func NewLoopRaw(processor func([]*item.BlockTxRaw) error) *LoopRaw {
	return &LoopRaw{
		Processor: processor,
	}
}
