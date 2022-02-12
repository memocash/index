package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"runtime"
)

type LockHeight struct {
	Verbose bool
}

func (t *LockHeight) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block for lock height")
	}
	if err := t.QueueTxs(block); err != nil {
		return jerr.Get("error queueing msg txs for lock height", err)
	}
	return nil
}

func (t *LockHeight) QueueTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block for lock height queue txs")
	}
	blockHash := block.BlockHash()
	var blockHashBytes []byte
	var height int64
	if !block.Header.Timestamp.IsZero() {
		blockHashBytes = blockHash.CloneBytes()
		blockHeight, err := item.GetBlockHeight(blockHashBytes)
		if err != nil {
			return jerr.Get("error getting block height for lock height", err)
		}
		height = blockHeight.Height
	}
	var objects []item.Object
	var lockHeightOutputsToRemove []*item.LockHeightOutput
	var objectCount int
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			/*objects = append(objects, &item.LockHeightOutputInput{
				LockHash: ,
				Hash:      txHashBytes,
				Index:     uint32(j),
				PrevHash:  tx.TxIn[j].PreviousOutPoint.Hash.CloneBytes(),
				PrevIndex: tx.TxIn[j].PreviousOutPoint.Index,
			})*/
		}
		for h := range tx.TxOut {
			lockHash := script.GetLockHash(tx.TxOut[h].PkScript)
			var lockHeightOutput = &item.LockHeightOutput{
				LockHash: lockHash,
				Height:   height,
				Hash:     txHashBytes,
				Index:    uint32(h),
			}
			if height > 0 {
				lockHeightOutputsToRemove = append(lockHeightOutputsToRemove, &item.LockHeightOutput{
					LockHash: lockHash,
					Hash:     txHashBytes,
					Index:    uint32(h),
				})
			} else {
				lockHeightOutput.Height = item.HeightMempool
			}
			objects = append(objects, lockHeightOutput)
			if len(objects) >= 10000 {
				if err := item.Save(objects); err != nil {
					return jerr.Get("error saving db lock height objects (at limit)", err)
				}
				objectCount += len(objects)
				objects = nil
				runtime.GC()
			}
		}
	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving db lock height objects", err)
	}
	if err := item.RemoveLockHeightOutputs(lockHeightOutputsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height outputs for lock heights", err)
	}
	objectCount += len(objects)
	jlog.Logf("Saved %d lock height objects, height: %d\n", objectCount, height)
	return nil
}

func NewLockHeight(verbose bool) *LockHeight {
	return &LockHeight{
		Verbose: verbose,
	}
}
