package saver

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/dbi"
)

type ClearMempoolTxRaw struct {
}

func (r *ClearMempoolTxRaw) SaveTxs(b *dbi.Block) error {
	var mempoolTxRawsToRemove = make([]*item.MempoolTxRaw, len(b.Transactions))
	for i := range b.Transactions {
		mempoolTxRawsToRemove[i] = &item.MempoolTxRaw{TxHash: b.Transactions[i].Hash[:]}
	}
	if err := item.RemoveMempoolTxRaws(mempoolTxRawsToRemove); err != nil {
		return jerr.Get("error removing mempool tx raws", err)
	}
	return nil
}

func NewClearMempoolTxRaw() *ClearMempoolTxRaw {
	return &ClearMempoolTxRaw{}
}
