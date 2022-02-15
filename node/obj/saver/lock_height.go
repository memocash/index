package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
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
	if height == 0 {
		height = item.HeightMempool
	}
	var objects []item.Object
	var lockHeightOutputsToRemove []*item.LockHeightOutput
	var objectCount int
	var outsToQueryInputs []memo.Out
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		for h := range tx.TxOut {
			var index = uint32(h)
			lockHash := script.GetLockHash(tx.TxOut[h].PkScript)
			var lockHeightOutput = &item.LockHeightOutput{
				LockHash: lockHash,
				Height:   height,
				Hash:     txHashBytes,
				Index:    index,
			}
			outsToQueryInputs = append(outsToQueryInputs, memo.Out{
				LockHash: lockHash,
				TxHash:   txHashBytes,
				Index:    index,
			})
			if height > 0 {
				lockHeightOutputsToRemove = append(lockHeightOutputsToRemove, &item.LockHeightOutput{
					LockHash: lockHash,
					Hash:     txHashBytes,
					Index:    index,
				})
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
	var outputsToGet []memo.Out
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			outputsToGet = append(outputsToGet, memo.Out{
				TxHash: txHashBytes,
				Index:  uint32(j),
			})
		}
	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving db lock height objects 1", err)
	}
	objects = nil
	outputs, err := item.GetTxOutputs(outputsToGet)
	if err != nil {
		return jerr.Get("error getting outputs for lock height inputs", err)
	}
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		for j := range tx.TxIn {
			var index = uint32(j)
			var lockHash []byte
			for _, output := range outputs {
				if bytes.Equal(output.TxHash, txHashBytes) && output.Index == index {
					lockHash = output.LockHash
					break
				}
			}
			objects = append(objects, &item.LockHeightOutputInput{
				LockHash:  lockHash,
				Height:    height,
				Hash:      txHashBytes,
				Index:     index,
				PrevHash:  tx.TxIn[j].PreviousOutPoint.Hash.CloneBytes(),
				PrevIndex: tx.TxIn[j].PreviousOutPoint.Index,
			})
		}
	}
	outputInputs, err := item.GetOutputInputs(outsToQueryInputs)
	if err != nil {
		return jerr.Get("error getting output inputs for lock output inputs", err)
	}
	var txHashes = make([][]byte, len(outputInputs))
	for i := range outputInputs {
		txHashes[i] = outputInputs[i].Hash
	}
	txHashes = jutil.RemoveDupesAndEmpties(txHashes)
	txBlocks, err := item.GetTxBlocks(txHashes)
	if err != nil {
		return jerr.Get("error getting tx blocks for lock height output inputs", err)
	}
	var blockHashesToGetHeights = make([][]byte, len(txBlocks))
	for i := range txBlocks {
		blockHashesToGetHeights[i] = txBlocks[i].BlockHash
	}
	blockHeights, err := item.GetBlockHeights(blockHashesToGetHeights)
	if err != nil {
		return jerr.Get("error getting block heights for lock height output inputs", err)
	}
	for _, outputInput := range outputInputs {
		var lockHash []byte
		for _, outsToQueryInput := range outsToQueryInputs {
			if bytes.Equal(outsToQueryInput.TxHash, outputInput.PrevHash) &&
				outsToQueryInput.Index == outputInput.PrevIndex {
				lockHash = outsToQueryInput.LockHash
				break
			}
		}
		var txBlockHash []byte
		for _, txBlock := range txBlocks {
			if bytes.Equal(txBlock.TxHash, outputInput.Hash) {
				txBlockHash = txBlock.BlockHash
				break
			}
		}
		var txBlockHeight int64
		if len(txBlockHash) > 0 {
			for _, blockHeight := range blockHeights {
				if bytes.Equal(blockHeight.BlockHash, txBlockHash) {
					txBlockHeight = blockHeight.Height
				}
			}
		}
		objects = append(objects, &item.LockHeightOutputInput{
			LockHash:  lockHash,
			Height:    txBlockHeight,
			Hash:      outputInput.Hash,
			Index:     outputInput.Index,
			PrevHash:  outputInput.PrevHash,
			PrevIndex: outputInput.PrevIndex,
		})
	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving db lock height objects 2", err)
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
