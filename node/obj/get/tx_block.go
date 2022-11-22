package get

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item/chain"
)

type TxBlock struct {
	Txs []*chain.TxBlock
}

func (b *TxBlock) Get(txHashes [][]byte) error {
	var err error
	b.Txs, err = chain.GetTxBlocks(txHashes)
	if err != nil {
		return jerr.Get("error getting tx blocks from queue", err)
	}
	return nil
}

func NewTxBlock() *TxBlock {
	return &TxBlock{}
}
