package block_tx

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
)

type Loop struct {
	Processor func([]*chain.BlockTx) error
}

func (l *Loop) Process(blockHash []byte) error {
	const limit = client.DefaultLimit
	var startUid []byte
	for {
		blockTxes, err := chain.GetBlockTxes(chain.BlockTxesRequest{
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
		startUid = chain.GetBlockTxUid(blockHash, blockTxes[len(blockTxes)-1].TxHash[:])
	}
	return nil
}

func NewLoop(processor func([]*chain.BlockTx) error) *Loop {
	return &Loop{
		Processor: processor,
	}
}
