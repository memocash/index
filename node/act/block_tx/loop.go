package block_tx

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
)

type Loop struct {
	Processor func([]*item.BlockTx) error
}

func (l *Loop) Process(blockHash []byte) error {
	const limit = client.DefaultLimit
	var startUid []byte
	for {
		blockTxes, err := item.GetBlockTxes(item.BlockTxesRequest{
			BlockHash: blockHash,
			StartUid:  startUid,
			Limit:     limit,
		})
		if err != nil {
			return jerr.Get("error getting block txs for loop process", err)
		}
		if err := l.Processor(blockTxes); err != nil {
			return jerr.Get("error processing block txes", err)
		}
		if len(blockTxes) < limit {
			break
		}
		startUid = item.GetBlockTxUid(blockHash, blockTxes[len(blockTxes)-1].TxHash)
	}
	return nil
}

func NewLoop(processor func([]*item.BlockTx) error) *Loop {
	return &Loop{
		Processor: processor,
	}
}
