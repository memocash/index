package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

type TxMinimal struct {
	Verbose bool
}

func (t *TxMinimal) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	if err := t.QueueTxs(*block); err != nil {
		return jerr.Get("error queueing msg txs", err)
	}
	return nil
}

func (t *TxMinimal) QueueTxs(block wire.MsgBlock) error {
	blockHash := block.BlockHash()
	var objects []db.Object
	for i, tx := range block.Transactions {
		txHash := tx.TxHash()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		if !block.Header.Timestamp.IsZero() {
			var blockTx = &chain.BlockTx{
				BlockHash: blockHash,
				Index:     uint32(i),
				TxHash:    txHash,
			}
			objects = append(objects, blockTx)
			objects = append(objects, &chain.TxBlock{
				TxHash:    txHash,
				BlockHash: blockHash,
				Index:     uint32(i),
			})
		}
		objects = append(objects, &chain.Tx{
			TxHash:   txHash,
			Version:  tx.Version,
			LockTime: tx.LockTime,
		})
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			objects = append(objects, &chain.TxInput{
				TxHash:       txHash,
				Index:        uint32(j),
				PrevHash:     tx.TxIn[j].PreviousOutPoint.Hash,
				PrevIndex:    tx.TxIn[j].PreviousOutPoint.Index,
				Sequence:     tx.TxIn[j].Sequence,
				UnlockScript: tx.TxIn[j].SignatureScript,
			})
			objects = append(objects, &chain.OutputInput{
				PrevHash:  tx.TxIn[j].PreviousOutPoint.Hash,
				PrevIndex: tx.TxIn[j].PreviousOutPoint.Index,
				Hash:      txHash,
				Index:     uint32(j),
			})
		}
		for k := range tx.TxOut {
			objects = append(objects, &chain.TxOutput{
				TxHash:     txHash,
				Index:      uint32(k),
				Value:      tx.TxOut[k].Value,
				LockScript: tx.TxOut[k].PkScript,
			})
		}
		objects = append(objects, &item.TxSeen{
			TxHash:    txHash[:],
			Timestamp: time.Now(),
		})
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db tx objects", err)
	}
	return nil
}

func NewTxMinimal(verbose bool) *TxMinimal {
	return &TxMinimal{
		Verbose: verbose,
	}
}
