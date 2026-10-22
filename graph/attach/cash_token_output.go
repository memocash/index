package attach

import (
	"context"
	"fmt"

	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type CashTokenOutputs struct {
	base
	Outputs []*model.CashTokenOutput
}

func ToCashTokenOutputs(ctx context.Context, fields []Field, outputs []*model.CashTokenOutput) error {
	if len(outputs) == 0 {
		return nil
	}
	o := CashTokenOutputs{
		base:    base{Ctx: ctx, Fields: fields},
		Outputs: outputs,
	}
	o.Wait.Add(1)
	go o.AttachOutputs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to cash token outputs; %w", o.Errors[0])
	}
	return nil
}

func (o *CashTokenOutputs) getOutputIndexMap() map[outKey][]int {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	m := make(map[outKey][]int, len(o.Outputs))
	for i := range o.Outputs {
		k := outKey{o.Outputs[i].Hash, o.Outputs[i].Index}
		m[k] = append(m[k], i)
	}
	return m
}

func (o *CashTokenOutputs) AttachOutputs() {
	defer o.Wait.Done()
	if !o.HasField([]string{"output"}) {
		return
	}
	o.Mutex.Lock()
	var outs = make([]memo.Out, 0, len(o.Outputs))
	for i := range o.Outputs {
		outs = append(outs, memo.Out{TxHash: o.Outputs[i].Hash[:], Index: o.Outputs[i].Index})
	}
	o.Mutex.Unlock()
	txOutputs, err := chain.GetTxOutputs(o.Ctx, outs)
	if err != nil {
		o.AddError(fmt.Errorf("error getting tx outputs for cash token outputs; %w", err))
		return
	}
	outputIndexMap := o.getOutputIndexMap()
	var allOutputs []*model.TxOutput
	o.Mutex.Lock()
	for j := range txOutputs {
		indices, ok := outputIndexMap[outKey{txOutputs[j].TxHash, txOutputs[j].Index}]
		if !ok {
			continue
		}
		for _, i := range indices {
			o.Outputs[i].Output = &model.TxOutput{
				Hash:   txOutputs[j].TxHash,
				Index:  txOutputs[j].Index,
				Amount: txOutputs[j].Value,
				Script: txOutputs[j].LockScript,
			}
			allOutputs = append(allOutputs, o.Outputs[i].Output)
		}
	}
	o.Mutex.Unlock()
	if err := ToOutputs(o.Ctx, GetPrefixFields(o.Fields, "output."), allOutputs); err != nil {
		o.AddError(fmt.Errorf("error attaching to outputs for cash token outputs; %w", err))
		return
	}
}
