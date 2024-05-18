package load

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type SlpBatons struct {
	baseA
	SlpBatons []*model.SlpBaton
}

func AttachToSlpBatons(ctx context.Context, fields []Field, slpBatons []*model.SlpBaton) error {
	if len(slpBatons) == 0 {
		return nil
	}
	o := SlpBatons{
		baseA:     baseA{Ctx: ctx, Fields: fields},
		SlpBatons: slpBatons,
	}
	o.Wait.Add(2)
	go o.AttachGeneses()
	go o.AttachOutputs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to slp batons; %w", o.Errors[0])
	}
	return nil
}

func (o *SlpBatons) GetTokenHashes() [][32]byte {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var tokenHashes [][32]byte
	for i := range o.SlpBatons {
		tokenHashes = append(tokenHashes, o.SlpBatons[i].TokenHash)
	}
	return tokenHashes
}

func (o *SlpBatons) GetOuts() []memo.Out {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var txOuts []memo.Out
	for i := range o.SlpBatons {
		txOuts = append(txOuts, memo.Out{
			TxHash: o.SlpBatons[i].Hash[:],
			Index:  o.SlpBatons[i].Index,
		})
	}
	return txOuts
}

func (o *SlpBatons) AttachGeneses() {
	defer o.Wait.Done()
	if !o.HasField([]string{"genesis"}) {
		return
	}
	slpGeneses, err := slp.GetGeneses(o.Ctx, o.GetTokenHashes())
	if err != nil {
		o.AddError(fmt.Errorf("error getting slp geneses from dataloader; %w", err))
		return
	}
	var allSlpGeneses []*model.SlpGenesis
	o.Mutex.Lock()
	for i := range o.SlpBatons {
		for j := range slpGeneses {
			if o.SlpBatons[i].TokenHash != slpGeneses[j].TxHash {
				continue
			}
			o.SlpBatons[i].Genesis = &model.SlpGenesis{
				Hash:       slpGeneses[j].TxHash,
				TokenType:  model.Uint8(slpGeneses[j].TokenType),
				Decimals:   model.Uint8(slpGeneses[j].Decimals),
				BatonIndex: slpGeneses[j].BatonIndex,
				Ticker:     slpGeneses[j].Ticker,
				Name:       slpGeneses[j].Name,
				DocURL:     slpGeneses[j].DocUrl,
				DocHash:    hex.EncodeToString(slpGeneses[j].DocHash[:]),
			}
			allSlpGeneses = append(allSlpGeneses, o.SlpBatons[i].Genesis)
			break
		}
	}
	o.Mutex.Unlock()
	if err := AttachToSlpGeneses(o.Ctx, GetPrefixFields(o.Fields, "genesis."), allSlpGeneses); err != nil {
		o.AddError(fmt.Errorf("error attaching to slp geneses for slp batons; %w", err))
		return
	}
}

func (o *SlpBatons) AttachOutputs() {
	defer o.Wait.Done()
	if !o.HasField([]string{"output"}) {
		return
	}
	txOutputs, err := chain.GetTxOutputs(o.Ctx, o.GetOuts())
	if err != nil {
		o.AddError(fmt.Errorf("error getting tx outputs for model slp batons; %w", err))
		return
	}
	var allOutputs []*model.TxOutput
	o.Mutex.Lock()
	for i := range o.SlpBatons {
		for j := range txOutputs {
			if o.SlpBatons[i].Hash != txOutputs[j].TxHash || o.SlpBatons[i].Index != txOutputs[j].Index {
				continue
			}
			o.SlpBatons[i].Output = &model.TxOutput{
				Hash:   txOutputs[j].TxHash,
				Index:  txOutputs[j].Index,
				Amount: txOutputs[j].Value,
				Script: txOutputs[j].LockScript,
			}
			allOutputs = append(allOutputs, o.SlpBatons[i].Output)
			break
		}
	}
	o.Mutex.Unlock()
	if err := AttachToOutputs(o.Ctx, GetPrefixFields(o.Fields, "output."), allOutputs); err != nil {
		o.AddError(fmt.Errorf("error attaching to outputs for slp batons; %w", err))
		return
	}
}
