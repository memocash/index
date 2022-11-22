package get

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item/chain"
)

type BlockTx struct {
	BlockHash []byte
	TxHash    []byte
	BlockTx   *chain.BlockTx
	TxBlock   *chain.TxBlock
}

func (b *BlockTx) Get() error {
	var err error
	b.BlockTx, err = chain.GetBlockTx(b.BlockHash, b.TxHash)
	if err != nil {
		return jerr.Get("error getting block tx from queue", err)
	}
	b.TxBlock, err = chain.GetSingleTxBlock(b.TxHash, b.BlockHash)
	if err != nil {
		return jerr.Get("error getting tx block from queue", err)
	}
	return nil
}

func NewBlockTx(blockHash, txHash []byte) *BlockTx {
	return &BlockTx{
		BlockHash: blockHash,
		TxHash:    txHash,
	}
}
