package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/status"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"runtime"
	"time"
)

type Tx struct {
	Verbose bool
	Shard   int
}

func (t *Tx) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	err := t.QueueTxs(block)
	if err != nil {
		return jerr.Get("error queueing msg txs", err)
	}
	return nil
}

func (t *Tx) QueueTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block or empty header")
	}
	blockHash := block.BlockHash()
	var blockHashBytes []byte
	if !block.Header.Timestamp.IsZero() {
		blockHashBytes = blockHash.CloneBytes()
	}
	var objects []db.Object
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		if len(blockHashBytes) > 0 {
			objects = append(objects, &item.BlockTx{
				TxHash:    txHashBytes,
				BlockHash: blockHashBytes,
			})
			objects = append(objects, &item.TxBlock{
				TxHash:    txHashBytes,
				BlockHash: blockHashBytes,
			})
		}
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			objects = append(objects, &item.OutputInput{
				Hash:      txHashBytes,
				Index:     uint32(j),
				PrevHash:  tx.TxIn[j].PreviousOutPoint.Hash.CloneBytes(),
				PrevIndex: tx.TxIn[j].PreviousOutPoint.Index,
			})
		}
		for h := range tx.TxOut {
			objects = append(objects, &item.LockOutput{
				LockHash: script.GetLockHash(tx.TxOut[h].PkScript),
				Hash:     txHashBytes,
				Index:    uint32(h),
			})
			if len(objects) >= 10000 {
				if err := db.Save(objects); err != nil {
					return jerr.Get("error saving db tx objects (at limit)", err)
				}
				objects = nil
				runtime.GC()
			}
		}
		objects = append(objects, &item.TxProcessed{
			TxHash:    txHashBytes,
			Timestamp: time.Now(),
		})
	}
	if len(blockHashBytes) > 0 {
		blockHeight, err := item.GetBlockHeight(blockHashBytes)
		if err != nil {
			return jerr.Get("error getting block height for tx save", err)
		}
		if t.Shard != status.NoShard {
			objects = append(objects, &item.HeightBlockShard{
				Shard:     uint(t.Shard),
				Height:    blockHeight.Height,
				BlockHash: blockHeight.BlockHash,
			})
		}
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db tx objects", err)
	}
	return nil
}

func NewTx(verbose bool) *Tx {
	return &Tx{
		Verbose: verbose,
	}
}

func NewTxShard(verbose bool, shard int) *Tx {
	return &Tx{
		Verbose: verbose,
		Shard:   shard,
	}
}
