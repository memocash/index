package get

import (
	"fmt"
	"github.com/memocash/index/db/item/chain"
)

type BlockTx struct {
	BlockHash [32]byte
	TxHash    [32]byte
	BlockTx   *chain.BlockTx
	TxBlock   *chain.TxBlock
}

func (b *BlockTx) Get() error {
	var err error
	b.TxBlock, err = chain.GetSingleTxBlock(b.TxHash, b.BlockHash)
	if err != nil {
		return fmt.Errorf("error getting tx block from queue; %w", err)
	}
	b.BlockTx, err = chain.GetBlockTx(b.BlockHash, b.TxBlock.Index)
	if err != nil {
		return fmt.Errorf("error getting block tx from queue; %w", err)
	}
	return nil
}

func NewBlockTx(blockHash, txHash [32]byte) *BlockTx {
	return &BlockTx{
		BlockHash: blockHash,
		TxHash:    txHash,
	}
}
