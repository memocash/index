package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
)

type ClearMempoolTxRaw struct {
}

func (r *ClearMempoolTxRaw) SaveTxs(block *wire.MsgBlock) error {
	var mempoolTxRawsToRemove = make([]*item.MempoolTxRaw, len(block.Transactions))
	for i := range block.Transactions {
		txHash := block.Transactions[i].TxHash()
		mempoolTxRawsToRemove[i] = &item.MempoolTxRaw{TxHash: txHash.CloneBytes()}
	}
	if err := item.RemoveMempoolTxRaws(mempoolTxRawsToRemove); err != nil {
		return jerr.Get("error removing mempool tx raws", err)
	}
	return nil
}

func NewClearMempoolTxRaw() *ClearMempoolTxRaw {
	return &ClearMempoolTxRaw{}
}
