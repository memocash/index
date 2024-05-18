package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Inputs struct {
	baseA
	Inputs []*model.TxInput
}

func AttachToInputs(ctx context.Context, fields []Field, inputs []*model.TxInput) error {
	i := Inputs{
		baseA:  baseA{Ctx: ctx, Fields: fields},
		Inputs: inputs,
	}
	i.Wait.Add(3)
	go i.AttachScriptSequence()
	go i.AttachTxs()
	go i.AttachTxOutputs()
	i.Wait.Wait()
	if len(i.Errors) > 0 {
		return fmt.Errorf("error attaching to inputs; %w", i.Errors[0])
	}
	return nil
}

func (i *Inputs) AttachScriptSequence() {
	defer i.Wait.Done()
	if !i.HasField([]string{"script", "sequence"}) {
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
	txInputs, err := chain.GetTxInputs(i.Ctx, outs)
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
	if !i.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	i.Mutex.Lock()
	for j := range i.Inputs {
		i.Inputs[j].Tx = &model.Tx{Hash: i.Inputs[j].Hash}
		allTxs = append(allTxs, i.Inputs[j].Tx)
	}
	i.Mutex.Unlock()
	if err := AttachToTxs(i.Ctx, GetPrefixFields(i.Fields, "tx."), allTxs); err != nil {
		i.AddError(fmt.Errorf("error attaching to txs for model tx inputs; %w", err))
		return
	}
}

func (i *Inputs) AttachTxOutputs() {
	defer i.Wait.Done()
	if !i.HasField([]string{"output"}) {
		return
	}
	var allOutputs []*model.TxOutput
	i.Mutex.Lock()
	for j := range i.Inputs {
		i.Inputs[j].Output = &model.TxOutput{Hash: i.Inputs[j].PrevHash, Index: i.Inputs[j].PrevIndex}
		allOutputs = append(allOutputs, i.Inputs[j].Output)
	}
	i.Mutex.Unlock()
	if err := AttachToOutputs(i.Ctx, GetPrefixFields(i.Fields, "output."), allOutputs); err != nil {
		i.AddError(fmt.Errorf("error attaching all to input tx output; %w", err))
		return
	}
}
