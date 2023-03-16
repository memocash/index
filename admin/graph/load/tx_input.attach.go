package load

import (
	"fmt"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Inputs struct {
	baseA
	Inputs []*model.TxInput
}

func AttachToInputs(preloads []string, inputs []*model.TxInput) error {
	i := Inputs{
		baseA:  baseA{Preloads: preloads},
		Inputs: inputs,
	}
	i.Wait.Add(2)
	go i.AttachScriptSequence()
	go i.AttachTxs()
	i.Wait.Wait()
	if len(i.Errors) > 0 {
		return fmt.Errorf("error attaching to inputs; %w", i.Errors[0])
	}
	return nil
}

func (i *Inputs) AttachScriptSequence() {
	defer i.Wait.Done()
	if !i.HasPreload([]string{"script", "sequence"}) {
		return
	}
	var outs []memo.Out
	i.Mutex.Lock()
	for j := range i.Inputs {
		if len(i.Inputs[j].Script) > 0 && i.Inputs[j].Sequence > 0 {
			continue
		}
		outs = append(outs, memo.Out{
			TxHash: i.Inputs[j].Hash[:],
			Index:  i.Inputs[j].Index,
		})
	}
	i.Mutex.Unlock()
	if len(outs) == 0 {
		return
	}
	txInputs, err := chain.GetTxInputs(outs)
	if err != nil {
		i.AddError(fmt.Errorf("error getting tx inputs for model tx inputs script sequence; %w", err))
		return
	}
	i.Mutex.Lock()
	defer i.Mutex.Unlock()
	for j := range i.Inputs {
		for k := range txInputs {
			if i.Inputs[j].Hash != txInputs[k].TxHash || i.Inputs[j].Index != txInputs[k].Index {
				continue
			}
			i.Inputs[j].Script = txInputs[k].UnlockScript
			i.Inputs[j].Sequence = txInputs[k].Sequence
			break
		}
	}
}

func (i *Inputs) AttachTxs() {
	defer i.Wait.Done()
	var txHashes = make([][32]byte, len(i.Inputs))
	i.Mutex.Lock()
	for j := range i.Inputs {
		txHashes[j] = i.Inputs[j].Hash
	}
	i.Mutex.Unlock()
	txs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		i.AddError(fmt.Errorf("error getting txs for model tx inputs; %w", err))
		return
	}
	var allTxs []*model.Tx
	i.Mutex.Lock()
	for j := range i.Inputs {
		for k := range txs {
			if i.Inputs[j].Hash != txs[k].TxHash {
				continue
			}
			i.Inputs[j].Tx = &model.Tx{
				Hash:     txs[k].TxHash,
				Version:  txs[k].Version,
				LockTime: txs[k].LockTime,
			}
			allTxs = append(allTxs, i.Inputs[j].Tx)
			break
		}
	}
	preloads := GetPrefixPreloads(i.Preloads, "tx.")
	i.Mutex.Unlock()
	if err := AttachToTxs(preloads, allTxs); err != nil {
		i.AddError(fmt.Errorf("error attaching to txs for model tx inputs; %w", err))
		return
	}
}
