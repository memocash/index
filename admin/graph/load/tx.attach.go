package load

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"sort"
	"sync"
	"time"
)

type Tx struct {
	baseA
	Txs         []*model.Tx
	DetailsWait sync.WaitGroup
}

func AttachToTxs(ctx context.Context, fields []Field, txs []*model.Tx) error {
	t := Tx{
		baseA: baseA{Ctx: ctx, Fields: fields},
		Txs:   txs,
	}
	t.DetailsWait.Add(3)
	t.Wait.Add(5)
	go t.AttachInputs()
	go t.AttachOutputs()
	go t.AttachInfo()
	go t.AttachSeens()
	go t.AttachBlocks()
	t.DetailsWait.Wait()
	go t.AttachRaws()
	t.Wait.Wait()
	if len(t.Errors) > 0 {
		return fmt.Errorf("error attaching details to txs; %w", t.Errors[0])
	}
	return nil
}

func (t *Tx) GetTxHashes(checkVersion, checkSeen bool) [][32]byte {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	var txHashes [][32]byte
	for i := range t.Txs {
		if checkVersion && t.Txs[i].Version != 0 {
			continue
		} else if checkSeen && !jutil.IsTimeZero(time.Time(t.Txs[i].Seen)) {
			continue
		}
		txHashes = append(txHashes, t.Txs[i].Hash)
	}
	return txHashes
}

func (t *Tx) AttachInputs() {
	defer t.DetailsWait.Done()
	if !t.HasField([]string{"inputs", "raw"}) {
		return
	}
	txHashes := t.GetTxHashes(false, false)
	txInputs, err := chain.GetTxInputsByHashes(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting tx inputs for model tx; %w", err))
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
		for j := range txInputs {
			if t.Txs[i].Hash != txInputs[j].TxHash {
				continue
			}
			t.Txs[i].Inputs = append(t.Txs[i].Inputs, &model.TxInput{
				Hash:      txInputs[j].TxHash,
				Index:     txInputs[j].Index,
				PrevHash:  txInputs[j].PrevHash,
				PrevIndex: txInputs[j].PrevIndex,
				Sequence:  txInputs[j].Sequence,
				Script:    txInputs[j].UnlockScript,
			})
		}
		sort.Slice(t.Txs[i].Inputs, func(a, b int) bool {
			return t.Txs[i].Inputs[a].Index < t.Txs[i].Inputs[b].Index
		})
	}
}

func (t *Tx) AttachOutputs() {
	defer t.DetailsWait.Done()
	defer func() {
		go t.AttachToOutputs()
	}()
	if !t.HasField([]string{"outputs", "raw"}) {
		return
	}
	txHashes := t.GetTxHashes(false, false)
	txOutputs, err := chain.GetTxOutputsByHashes(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting tx outputs for model tx; %w", err))
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
		for j := range txOutputs {
			if t.Txs[i].Hash != txOutputs[j].TxHash {
				continue
			}
			t.Txs[i].Outputs = append(t.Txs[i].Outputs, &model.TxOutput{
				Hash:   txOutputs[j].TxHash,
				Index:  txOutputs[j].Index,
				Amount: txOutputs[j].Value,
				Script: txOutputs[j].LockScript,
			})
		}
		sort.Slice(t.Txs[i].Outputs, func(a, b int) bool {
			return t.Txs[i].Outputs[a].Index < t.Txs[i].Outputs[b].Index
		})
	}
}

func (t *Tx) AttachToOutputs() {
	defer t.Wait.Done()
	var allOutputs []*model.TxOutput
	t.Mutex.Lock()
	for _, tx := range t.Txs {
		allOutputs = append(allOutputs, tx.Outputs...)
	}
	prefixFields := GetPrefixFields(t.Fields, "outputs.")
	t.Mutex.Unlock()
	if err := AttachToOutputs(t.Ctx, prefixFields, allOutputs); err != nil {
		t.AddError(fmt.Errorf("error attaching to outputs for tx; %w", err))
		return
	}
}

func (t *Tx) AttachInfo() {
	defer t.DetailsWait.Done()
	if !t.HasField([]string{"version", "locktime", "raw"}) {
		return
	}
	txHashes := t.GetTxHashes(true, false)
	if len(txHashes) == 0 {
		return
	}
	chainTxs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting chain txs for raw; %w", err))
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
		for j := range chainTxs {
			if t.Txs[i].Hash != chainTxs[j].TxHash {
				continue
			}
			t.Txs[i].Version = chainTxs[j].Version
			t.Txs[i].LockTime = chainTxs[j].LockTime
			break
		}
	}
}

var TxMissingError = fmt.Errorf("error tx missing")

func (t *Tx) AttachRaws() {
	defer t.Wait.Done()
	if !t.HasField([]string{"raw"}) {
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
		if t.Txs[i].Version == 0 {
			t.AddError(fmt.Errorf("error tx missing info data: %s; %w", t.Txs[i].Hash, TxMissingError))
			return
		}
		var msgTx = &wire.MsgTx{
			Version:  t.Txs[i].Version,
			LockTime: t.Txs[i].LockTime,
		}
		for _, txIn := range t.Txs[i].Inputs {
			msgTx.TxIn = append(msgTx.TxIn, &wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash(txIn.PrevHash),
					Index: txIn.PrevIndex,
				},
				SignatureScript: txIn.Script,
				Sequence:        txIn.Sequence,
			})
		}
		for _, txOut := range t.Txs[i].Outputs {
			msgTx.TxOut = append(msgTx.TxOut, &wire.TxOut{
				Value:    txOut.Amount,
				PkScript: txOut.Script,
			})
		}
		if msgTx.TxHash() != chainhash.Hash(t.Txs[i].Hash) {
			t.AddError(fmt.Errorf("tx hash mismatch for raw: %s %s", msgTx.TxHash(), chainhash.Hash(t.Txs[i].Hash)))
			return
		}
		t.Txs[i].Raw = memo.GetRaw(msgTx)
	}
}

func (t *Tx) AttachSeens() {
	defer t.Wait.Done()
	if !t.HasField([]string{"seen"}) {
		return
	}
	txHashes := t.GetTxHashes(false, true)
	if len(txHashes) == 0 {
		return
	}
	txSeens, err := chain.GetTxSeens(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting chain txs for raw; %w", err))
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
		for j := range txSeens {
			if t.Txs[i].Hash != txSeens[j].TxHash {
				continue
			}
			t.Txs[i].Seen = model.Date(txSeens[j].Timestamp)
			break
		}
	}
}

func (t *Tx) AttachBlocks() {
	defer t.Wait.Done()
	defer func() {
		go t.AttachToBlocks()
	}()
	if !t.HasField([]string{"blocks"}) {
		return
	}
	txHashes := t.GetTxHashes(false, false)
	txBlocks, err := chain.GetTxBlocks(t.Ctx, txHashes)
	if err != nil {
		t.AddError(fmt.Errorf("error getting blocks for tx for block loader; %w", err))
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
		for j := range txBlocks {
			if t.Txs[i].Hash != txBlocks[j].TxHash {
				continue
			}
			t.Txs[i].Blocks = append(t.Txs[i].Blocks, &model.TxBlock{
				TxHash:    t.Txs[i].Hash,
				Tx:        t.Txs[i],
				BlockHash: txBlocks[j].BlockHash,
				Block:     &model.Block{Hash: txBlocks[j].BlockHash},
				Index:     txBlocks[j].Index,
			})
		}
		sort.Slice(t.Txs[i].Outputs, func(a, b int) bool {
			return t.Txs[i].Outputs[a].Index < t.Txs[i].Outputs[b].Index
		})
	}
}

func (t *Tx) AttachToBlocks() {
	defer t.Wait.Done()
	var allBlocks []*model.Block
	t.Mutex.Lock()
	for _, tx := range t.Txs {
		for _, txBlock := range tx.Blocks {
			allBlocks = append(allBlocks, txBlock.Block)
		}
	}
	prefixFields := GetPrefixFields(t.Fields, "blocks.block.")
	t.Mutex.Unlock()
	if err := AttachToBlocks(t.Ctx, prefixFields, allBlocks); err != nil {
		t.AddError(fmt.Errorf("error attaching to blocks for tx; %w", err))
		return
	}
}
