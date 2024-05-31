package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type SlpGeneses struct {
	base
	SlpGeneses []*model.SlpGenesis
}

func ToSlpGeneses(ctx context.Context, fields []Field, slpGeneses []*model.SlpGenesis) error {
	if len(slpGeneses) == 0 {
		return nil
	}
	o := SlpGeneses{
		base:       base{Ctx: ctx, Fields: fields},
		SlpGeneses: slpGeneses,
	}
	o.Wait.Add(3)
	go o.AttachSlpOutputs()
	go o.AttachSlpBatons()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to slp geneses; %w", o.Errors[0])
	}
	return nil
}

func (o *SlpGeneses) GetTokenOuts() []memo.Out {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var txOuts []memo.Out
	for i := range o.SlpGeneses {
		txOuts = append(txOuts, memo.Out{
			TxHash: o.SlpGeneses[i].Hash[:],
			Index:  memo.SlpMintTokenIndex,
		})
	}
	return txOuts
}

func (o *SlpGeneses) GetBatonOuts() []memo.Out {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var txOuts []memo.Out
	for i := range o.SlpGeneses {
		txOuts = append(txOuts, memo.Out{
			TxHash: o.SlpGeneses[i].Hash[:],
			Index:  o.SlpGeneses[i].BatonIndex,
		})
	}
	return txOuts
}

func (o *SlpGeneses) AttachSlpOutputs() {
	defer o.Wait.Done()
	if !o.HasField([]string{"output"}) {
		return
	}
	slpOutputs, err := slp.GetOutputs(o.Ctx, o.GetTokenOuts())
	if err != nil {
		o.AddError(fmt.Errorf("error getting tx outputs for model slp geneses; %w", err))
		return
	}
	var allSlpOutputs []*model.SlpOutput
	o.Mutex.Lock()
	for i := range o.SlpGeneses {
		for j := range slpOutputs {
			if o.SlpGeneses[i].Hash != slpOutputs[j].TxHash || memo.SlpMintTokenIndex != slpOutputs[j].Index {
				continue
			}
			o.SlpGeneses[i].Output = &model.SlpOutput{
				Hash:      slpOutputs[j].TxHash,
				Index:     slpOutputs[j].Index,
				TokenHash: slpOutputs[j].TokenHash,
				Amount:    slpOutputs[j].Quantity,
			}
			allSlpOutputs = append(allSlpOutputs, o.SlpGeneses[i].Output)
			break
		}
	}
	o.Mutex.Unlock()
	if err := ToSlpOutputs(o.Ctx, GetPrefixFields(o.Fields, "output."), allSlpOutputs); err != nil {
		o.AddError(fmt.Errorf("error attaching to slp outputs for slp geneses; %w", err))
		return
	}
}

func (o *SlpGeneses) AttachSlpBatons() {
	defer o.Wait.Done()
	if !o.HasField([]string{"baton"}) {
		return
	}
	batonOutputs, err := slp.GetBatons(o.Ctx, o.GetBatonOuts())
	if err != nil {
		o.AddError(fmt.Errorf("error getting tx outputs for model slp geneses; %w", err))
		return
	}
	var allBatons []*model.SlpBaton
	o.Mutex.Lock()
	for i := range o.SlpGeneses {
		for j := range batonOutputs {
			if o.SlpGeneses[i].Hash != batonOutputs[j].TxHash || o.SlpGeneses[i].BatonIndex != batonOutputs[j].Index {
				continue
			}
			o.SlpGeneses[i].Baton = &model.SlpBaton{
				Hash:      batonOutputs[j].TxHash,
				Index:     batonOutputs[j].Index,
				TokenHash: batonOutputs[j].TokenHash,
			}
			allBatons = append(allBatons, o.SlpGeneses[i].Baton)
			break
		}
	}
	o.Mutex.Unlock()
	if err := ToSlpBatons(o.Ctx, GetPrefixFields(o.Fields, "baton."), allBatons); err != nil {
		o.AddError(fmt.Errorf("error attaching to slp batons for slp geneses; %w", err))
		return
	}
}

func (o *SlpGeneses) AttachTxs() {
	defer o.Wait.Done()
	if !o.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	o.Mutex.Lock()
	for j := range o.SlpGeneses {
		o.SlpGeneses[j].Tx = &model.Tx{Hash: o.SlpGeneses[j].Hash}
		allTxs = append(allTxs, o.SlpGeneses[j].Tx)
	}
	o.Mutex.Unlock()
	if err := ToTxs(o.Ctx, GetPrefixFields(o.Fields, "tx."), allTxs); err != nil {
		o.AddError(fmt.Errorf("error attaching to txs for model slp geneses; %w", err))
		return
	}
}
