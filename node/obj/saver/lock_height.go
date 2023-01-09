package saver

import (
	"bytes"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/dbi"
	"runtime"
)

type LockHeight struct {
	Verbose     bool
	InitialSync bool
	CheckTxHash []byte
}

func (t *LockHeight) SaveTxs(b *dbi.Block) error {
	if b.IsNil() {
		return jerr.Newf("error nil block for lock height")
	}
	block := b.ToWireBlock()
	saveRun := NewLockHeightSaveRun(t.Verbose, t.InitialSync)
	saveRun.CheckTxHash = t.CheckTxHash
	if err := saveRun.SetHashHeightInOuts(block); err != nil {
		return jerr.Get("error setting hash height for lock height saver run", err)
	}
	if err := saveRun.SaveOutputs(); err != nil {
		return jerr.Get("error saving outputs for lock height saver run", err)
	}
	if err := saveRun.SaveOutputInputsForInputs(); err != nil {
		return jerr.Get("error saving output inputs for lock height saver run", err)
	}
	if err := saveRun.SaveOutputInputsForOutputs(); err != nil {
		return jerr.Get("error saving output inputs for outputs for lock height saver run", err)
	}
	var noLockHashStatus = ""
	if saveRun.NoLockHash > 0 {
		noLockHashStatus = fmt.Sprintf(" (%d)", saveRun.NoLockHash)
	}
	if t.Verbose {
		jlog.Logf("Saved %d%s lock height objects, height: %d\n", saveRun.ObjectCount, noLockHashStatus, saveRun.Height)
	}
	return nil
}

type LockHeightSaveRun struct {
	Verbose     bool
	InitialSync bool
	BlockHash   []byte
	Height      int64
	ObjectCount int
	NoLockHash  int
	Ins         []memo.InOut
	LockOuts    []memo.Out
	CheckTxHash []byte
}

func (t *LockHeightSaveRun) SetHashHeightInOuts(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block for lock height queue txs")
	}
	blockHash := block.BlockHash()
	if dbi.BlockHeaderSet(block.Header) {
		t.BlockHash = blockHash.CloneBytes()
		blockHeight, err := chain.GetBlockHeight(t.BlockHash)
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
		if t.Verbose || (len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, txHashBytes)) {
			jlog.Logf("lock height tx: %s\n", txHash.String())
		}
		if len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, txHashBytes) {
			jlog.Logf("Adding checked tx: %s (ins: %d, outs: %d)\n",
				hs.GetTxString(txHashBytes), len(tx.TxIn), len(tx.TxOut))
		}
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			t.Ins = append(t.Ins, memo.InOut{
				Hash:      txHashBytes,
				Index:     uint32(j),
				PrevHash:  tx.TxIn[j].PreviousOutPoint.Hash.CloneBytes(),
				PrevIndex: tx.TxIn[j].PreviousOutPoint.Index,
			})
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
	var objects []db.Object
	var lockHeightOutputsToRemove []*item.LockHeightOutput
	for _, lockOut := range t.LockOuts {
		var lockHeightOutput = &item.LockHeightOutput{
			LockHash: lockOut.LockHash,
			Height:   t.Height,
			Hash:     lockOut.TxHash,
			Index:    lockOut.Index,
		}
		if len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, lockHeightOutput.Hash) {
			jlog.Logf("Saving lock height output: %s:%d\n",
				hs.GetTxString(lockHeightOutput.Hash), lockHeightOutput.Index)
		}
		objects = append(objects, lockHeightOutput)
		if t.Height > 0 {
			lockHeightOutputsToRemove = append(lockHeightOutputsToRemove, &item.LockHeightOutput{
				LockHash: lockOut.LockHash,
				Height:   item.HeightMempool,
				Hash:     lockOut.TxHash,
				Index:    lockOut.Index,
			})
		}
		if len(objects) >= 10000 {
			if err := db.Save(objects); err != nil {
				return jerr.Get("error saving db lock height objects (at limit)", err)
			}
			t.ObjectCount += len(objects)
			objects = nil
			runtime.GC()
		}
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db lock height outputs", err)
	}
	t.ObjectCount += len(objects)
	if t.InitialSync {
		return nil
	}
	if err := item.RemoveLockHeightOutputs(lockHeightOutputsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height outputs for lock heights", err)
	}
	return nil
}

func (t *LockHeightSaveRun) SaveOutputInputsForInputs() error {
	var objects []db.Object
	var inputOuts []memo.Out
	var lockHeightOutputInputsToRemove []*item.LockHeightOutputInput
TxInLoop:
	for _, in := range t.Ins {
		for _, lockOut := range t.LockOuts {
			if bytes.Equal(lockOut.TxHash, in.PrevHash) && lockOut.Index == in.PrevIndex {
				continue TxInLoop
			}
		}
		inputOuts = append(inputOuts, memo.Out{
			TxHash: in.PrevHash,
			Index:  in.PrevIndex,
		})
	}
	inputOutputs, err := chain.GetTxOutputs(inputOuts)
	if err != nil {
		return jerr.Get("error getting outputs for lock height inputs", err)
	}
	if t.Verbose {
		jlog.Logf("inputOutputs: %d, inputOuts: %d\n", len(inputOutputs), len(inputOuts))
	}
	for _, in := range t.Ins {
		var lockHash []byte
		for _, inputOutput := range inputOutputs {
			if bytes.Equal(inputOutput.TxHash[:], in.PrevHash) && inputOutput.Index == in.PrevIndex {
				lockHash = script.GetLockHash(inputOutput.LockScript)
				break
			}
		}
		if lockHash == nil {
			for _, lockOut := range t.LockOuts {
				if bytes.Equal(lockOut.TxHash, in.PrevHash) && lockOut.Index == in.PrevIndex {
					lockHash = lockOut.LockHash
					break
				}
			}
		}
		if lockHash == nil {
			jlog.Logf("lock hash is nil for input (%s:%d) with output (%s:%d)\n",
				hs.GetTxString(in.Hash), in.Index, hs.GetTxString(in.PrevHash), in.PrevIndex)
			t.NoLockHash++
			continue
		}
		var lockHeightOutputInput = &item.LockHeightOutputInput{
			LockHash:  lockHash,
			Height:    t.Height,
			Hash:      in.Hash,
			Index:     in.Index,
			PrevHash:  in.PrevHash,
			PrevIndex: in.PrevIndex,
		}
		if len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, lockHeightOutputInput.Hash) {
			jlog.Logf("Saving lock height output input (for input): %s:%d\n",
				hs.GetTxString(lockHeightOutputInput.Hash), lockHeightOutputInput.Index)
		}
		if len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, lockHeightOutputInput.PrevHash) {
			jlog.Logf("Saving lock height output input (for input - prev): %s:%d\n",
				hs.GetTxString(lockHeightOutputInput.PrevHash), lockHeightOutputInput.PrevIndex)
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
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db lock height output inputs for inputs", err)
	}
	t.ObjectCount += len(objects)
	if t.InitialSync {
		return nil
	}
	if err := item.RemoveLockHeightOutputInputs(lockHeightOutputInputsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height output inputs for inputs for lock heights", err)
	}
	return nil
}

func (t *LockHeightSaveRun) SaveOutputInputsForOutputs() error {
	var objects []db.Object
	var lockHeightOutputInputsToRemove []*item.LockHeightOutputInput
	var lockOuts = t.LockOuts
	for _, in := range t.Ins {
		for i := 0; i < len(lockOuts); i++ {
			if bytes.Equal(lockOuts[i].TxHash, in.PrevHash) && lockOuts[i].Index == in.PrevIndex {
				lockOuts = append(lockOuts[:i], lockOuts[i+1:]...)
				i--
			}
		}
	}
	outputInputs, err := chain.GetOutputInputs(lockOuts)
	if err != nil {
		return jerr.Get("error getting output inputs for lock output inputs", err)
	}
	if t.Verbose {
		jlog.Logf("outputInputs: %d, lockOuts: %d\n", len(outputInputs), len(lockOuts))
	}
	var txHashes = make([][]byte, len(outputInputs))
	for i := range outputInputs {
		txHashes[i] = outputInputs[i].Hash[:]
	}
	txHashes = jutil.RemoveDupesAndEmpties(txHashes)
	txBlocks, err := chain.GetTxBlocks(db.RawTxHashesToFixed(txHashes))
	if err != nil {
		return jerr.Get("error getting tx blocks for lock height output inputs", err)
	}
	var blockHashesToGetHeights = make([][]byte, len(txBlocks))
	for i := range txBlocks {
		blockHashesToGetHeights[i] = txBlocks[i].BlockHash[:]
	}
	blockHeights, err := chain.GetBlockHeights(blockHashesToGetHeights)
	if err != nil {
		return jerr.Get("error getting block heights for lock height output inputs", err)
	}
	for _, outputInput := range outputInputs {
		var lockHash []byte
		for _, lockOut := range lockOuts {
			if bytes.Equal(lockOut.TxHash, outputInput.PrevHash[:]) &&
				lockOut.Index == outputInput.PrevIndex {
				lockHash = lockOut.LockHash
				break
			}
		}
		var txBlockHash *chainhash.Hash
		for _, txBlock := range txBlocks {
			if txBlock.TxHash == outputInput.Hash {
				blockHash := chainhash.Hash(txBlock.BlockHash)
				txBlockHash = &blockHash
				break
			}
		}
		var txBlockHeight int64
		if txBlockHash != nil {
			for _, blockHeight := range blockHeights {
				if blockHeight.BlockHash == *txBlockHash {
					txBlockHeight = blockHeight.Height
				}
			}
		}
		var lockHeightOutputInput = &item.LockHeightOutputInput{
			LockHash:  lockHash,
			Height:    txBlockHeight,
			Hash:      outputInput.Hash[:],
			Index:     outputInput.Index,
			PrevHash:  outputInput.PrevHash[:],
			PrevIndex: outputInput.PrevIndex,
		}
		if len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, lockHeightOutputInput.Hash) {
			jlog.Logf("Saving lock height output input (for output): %s:%d\n",
				hs.GetTxString(lockHeightOutputInput.Hash), lockHeightOutputInput.Index)
		}
		if len(t.CheckTxHash) > 0 && bytes.Equal(t.CheckTxHash, lockHeightOutputInput.PrevHash) {
			jlog.Logf("Saving lock height output input (for output - prev): %s:%d\n",
				hs.GetTxString(lockHeightOutputInput.PrevHash), lockHeightOutputInput.PrevIndex)
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
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db lock height output inputs for outputs", err)
	}
	t.ObjectCount += len(objects)
	if t.InitialSync {
		return nil
	}
	if err := item.RemoveLockHeightOutputInputs(lockHeightOutputInputsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height output inputs for outputs for lock heights", err)
	}
	return nil
}

func NewLockHeightSaveRun(verbose, initialSync bool) *LockHeightSaveRun {
	return &LockHeightSaveRun{
		Verbose:     verbose,
		InitialSync: initialSync,
	}
}

func NewLockHeight(verbose bool) *LockHeight {
	return &LockHeight{
		Verbose: verbose,
	}
}
