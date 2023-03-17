package load

import (
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

func AttachToTxs(preloads []string, txs []*model.Tx) error {
	t := Tx{
		baseA: baseA{Preloads: preloads},
		Txs:   txs,
	}
	t.DetailsWait.Add(3)
	t.Wait.Add(3)
	go t.AttachInputs()
	go t.AttachOutputs()
	go t.AttachInfo()
	go t.AttachSeens()
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
	if !t.HasPreload([]string{"inputs", "raw"}) {
		return
	}
	txHashes := t.GetTxHashes(false, false)
	txInputs, err := chain.GetTxInputsByHashes(txHashes)
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
	if !t.HasPreload([]string{"outputs", "raw"}) {
		return
	}
	txHashes := t.GetTxHashes(false, false)
	txOutputs, err := chain.GetTxOutputsByHashes(txHashes)
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
	preloads := GetPrefixPreloads(t.Preloads, "outputs.")
	t.Mutex.Unlock()
	if err := AttachToOutputs(preloads, allOutputs); err != nil {
		t.AddError(fmt.Errorf("error attaching to outputs for tx; %w", err))
		return
	}
}

func (t *Tx) AttachInfo() {
	defer t.DetailsWait.Done()
	if !t.HasPreload([]string{"version", "locktime", "raw"}) {
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

func (t *Tx) AttachRaws() {
	defer t.Wait.Done()
	if !t.HasPreload([]string{"raw"}) {
		return
	}
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	for i := range t.Txs {
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
	if !t.HasPreload([]string{"seen"}) {
		return
	}
	txHashes := t.GetTxHashes(false, true)
	if len(txHashes) == 0 {
		return
	}
	txSeens, err := chain.GetTxSeens(txHashes)
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
