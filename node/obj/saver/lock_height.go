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
	saveRun := NewLockHeightSaveRun(t.Verbose)
	if err := saveRun.SetHashHeightOuts(block); err != nil {
		return jerr.Get("error setting hash height for lock height saver run", err)
	}
	if err := saveRun.SaveOutputs(); err != nil {
		return jerr.Get("error saving outputs for lock height saver run", err)
	}
	if err := saveRun.SaveOutputInputsForInputs(block); err != nil {
		return jerr.Get("error saving output inputs for lock height saver run", err)
	}
	if err := saveRun.SaveOutputInputsForOutputs(); err != nil {
		return jerr.Get("error saving output inputs for outputs for lock height saver run", err)
	}
	jlog.Logf("Saved %d lock height objects, height: %d\n", saveRun.ObjectCount, saveRun.Height)
	return nil
}

type LockHeightSaveRun struct {
	Verbose     bool
	BlockHash   []byte
	Height      int64
	ObjectCount int
	LockOuts    []memo.Out
}

func (t *LockHeightSaveRun) SetHashHeightOuts(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block for lock height queue txs")
	}
	blockHash := block.BlockHash()
	if !block.Header.Timestamp.IsZero() {
		t.BlockHash = blockHash.CloneBytes()
		blockHeight, err := item.GetBlockHeight(t.BlockHash)
		if err != nil {
			return jerr.Get("error getting block height for lock height", err)
		}
		t.Height = blockHeight.Height
	}
	if t.Height == 0 {
		t.Height = item.HeightMempool
	}
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		for h := range tx.TxOut {
			lockHash := script.GetLockHash(tx.TxOut[h].PkScript)
			t.LockOuts = append(t.LockOuts, memo.Out{
				LockHash: lockHash,
				TxHash:   txHashBytes,
				Index:    uint32(h),
			})
		}
	}
	return nil
}

func (t *LockHeightSaveRun) SaveOutputs() error {
	var objects []item.Object
	var lockHeightOutputsToRemove []*item.LockHeightOutput
	for _, lockOut := range t.LockOuts {
		var lockHeightOutput = &item.LockHeightOutput{
			LockHash: lockOut.LockHash,
			Height:   t.Height,
			Hash:     lockOut.TxHash,
			Index:    lockOut.Index,
		}
		objects = append(objects, lockHeightOutput)
		if t.Height > 0 {
			lockHeightOutputsToRemove = append(lockHeightOutputsToRemove, &item.LockHeightOutput{
				LockHash: lockOut.LockHash,
				Hash:     lockOut.TxHash,
				Index:    lockOut.Index,
			})
		}
		if len(objects) >= 10000 {
			if err := item.Save(objects); err != nil {
				return jerr.Get("error saving db lock height objects (at limit)", err)
			}
			t.ObjectCount += len(objects)
			objects = nil
			runtime.GC()
		}

	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving db lock height outputs", err)
	}
	t.ObjectCount += len(objects)
	if err := item.RemoveLockHeightOutputs(lockHeightOutputsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height outputs for lock heights", err)
	}
	return nil
}

func (t *LockHeightSaveRun) SaveOutputInputsForInputs(block *wire.MsgBlock) error {
	var objects []item.Object
	var inputOuts []memo.Out
	var lockHeightOutputInputsToRemove []*item.LockHeightOutputInput
	for _, tx := range block.Transactions {
	TxInLoop:
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			var out = memo.Out{
				TxHash: tx.TxIn[j].PreviousOutPoint.Hash.CloneBytes(),
				Index:  tx.TxIn[j].PreviousOutPoint.Index,
			}
			for _, lockOut := range t.LockOuts {
				if bytes.Equal(lockOut.TxHash, out.TxHash) && lockOut.Index == out.Index {
					continue TxInLoop
				}
			}
			inputOuts = append(inputOuts, out)
		}
	}
	inputOutputs, err := item.GetTxOutputs(inputOuts)
	if err != nil {
		return jerr.Get("error getting outputs for lock height inputs", err)
	}
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			var index = uint32(j)
			var lockHash []byte
			for _, inputOutput := range inputOutputs {
				if bytes.Equal(inputOutput.TxHash, txHashBytes) && inputOutput.Index == index {
					lockHash = inputOutput.LockHash
					break
				}
			}
			if lockHash == nil {
				for _, lockOut := range t.LockOuts {
					if bytes.Equal(lockOut.TxHash, txHashBytes) && lockOut.Index == index {
						lockHash = lockOut.LockHash
						break
					}
				}
			}
			var lockHeightOutputInput = &item.LockHeightOutputInput{
				LockHash:  lockHash,
				Height:    t.Height,
				Hash:      txHashBytes,
				Index:     index,
				PrevHash:  tx.TxIn[j].PreviousOutPoint.Hash.CloneBytes(),
				PrevIndex: tx.TxIn[j].PreviousOutPoint.Index,
			}
			objects = append(objects, lockHeightOutputInput)
			if t.Height > 0 {
				lockHeightOutputInputsToRemove = append(lockHeightOutputInputsToRemove, &item.LockHeightOutputInput{
					LockHash:  lockHeightOutputInput.LockHash,
					Hash:      lockHeightOutputInput.Hash,
					Index:     lockHeightOutputInput.Index,
					PrevHash:  lockHeightOutputInput.PrevHash,
					PrevIndex: lockHeightOutputInput.PrevIndex,
				})
			}
		}
	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving db lock height output inputs for inputs", err)
	}
	if err := item.RemoveLockHeightOutputInputs(lockHeightOutputInputsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height output inputs for lock heights", err)
	}
	t.ObjectCount += len(objects)
	return nil
}

func (t *LockHeightSaveRun) SaveOutputInputsForOutputs() error {
	var objects []item.Object
	outputInputs, err := item.GetOutputInputs(t.LockOuts)
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
		for _, lockOut := range t.LockOuts {
			if bytes.Equal(lockOut.TxHash, outputInput.PrevHash) &&
				lockOut.Index == outputInput.PrevIndex {
				lockHash = lockOut.LockHash
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
	t.ObjectCount += len(objects)
	return nil
}

func NewLockHeightSaveRun(verbose bool) *LockHeightSaveRun {
	return &LockHeightSaveRun{
		Verbose: verbose,
	}
}

func NewLockHeight(verbose bool) *LockHeight {
	return &LockHeight{
		Verbose: verbose,
	}
}
