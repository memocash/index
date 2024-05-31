package attach

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type SlpOutputs struct {
	base
	SlpOutputs []*model.SlpOutput
}

func ToSlpOutputs(ctx context.Context, fields []Field, slpOutputs []*model.SlpOutput) error {
	if len(slpOutputs) == 0 {
		return nil
	}
	o := SlpOutputs{
		base:       base{Ctx: ctx, Fields: fields},
		SlpOutputs: slpOutputs,
	}
	o.Wait.Add(2)
	go o.AttachGeneses()
	go o.AttachOutputs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to slp outputs; %w", o.Errors[0])
	}
	return nil
}

func (o *SlpOutputs) GetTokenHashes() [][32]byte {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var tokenHashes [][32]byte
	for i := range o.SlpOutputs {
		tokenHashes = append(tokenHashes, o.SlpOutputs[i].TokenHash)
	}
	return tokenHashes
}

func (o *SlpOutputs) GetOuts() []memo.Out {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	var txOuts []memo.Out
	for i := range o.SlpOutputs {
		txOuts = append(txOuts, memo.Out{
			TxHash: o.SlpOutputs[i].Hash[:],
			Index:  o.SlpOutputs[i].Index,
		})
	}
	return txOuts
}

func (o *SlpOutputs) AttachGeneses() {
	defer o.Wait.Done()
	if !o.HasField([]string{"genesis"}) {
		return
	}
	slpGeneses, err := slp.GetGeneses(o.Ctx, o.GetTokenHashes())
	if err != nil {
		o.AddError(fmt.Errorf("error getting slp geneses for attach to slp outputs; %w", err))
		return
	}
	var allSlpGeneses []*model.SlpGenesis
	o.Mutex.Lock()
	for i := range o.SlpOutputs {
		for j := range slpGeneses {
			if o.SlpOutputs[i].TokenHash != slpGeneses[j].TxHash {
				continue
			}
			o.SlpOutputs[i].Genesis = &model.SlpGenesis{
				Hash:       slpGeneses[j].TxHash,
				TokenType:  model.Uint8(slpGeneses[j].TokenType),
				Decimals:   model.Uint8(slpGeneses[j].Decimals),
				BatonIndex: slpGeneses[j].BatonIndex,
				Ticker:     slpGeneses[j].Ticker,
				Name:       slpGeneses[j].Name,
				DocURL:     slpGeneses[j].DocUrl,
				DocHash:    hex.EncodeToString(slpGeneses[j].DocHash[:]),
			}
			allSlpGeneses = append(allSlpGeneses, o.SlpOutputs[i].Genesis)
			break
		}
	}
	o.Mutex.Unlock()
	if err := ToSlpGeneses(o.Ctx, GetPrefixFields(o.Fields, "genesis."), allSlpGeneses); err != nil {
		o.AddError(fmt.Errorf("error attaching to slp geneses for slp outputs; %w", err))
		return
	}
}

func (o *SlpOutputs) AttachOutputs() {
	defer o.Wait.Done()
	if !o.HasField([]string{"output"}) {
		return
	}
	txOutputs, err := chain.GetTxOutputs(o.Ctx, o.GetOuts())
	if err != nil {
		o.AddError(fmt.Errorf("error getting tx outputs for model slp outputs; %w", err))
		return
	}
	var allOutputs []*model.TxOutput
	o.Mutex.Lock()
	for i := range o.SlpOutputs {
		for j := range txOutputs {
			if o.SlpOutputs[i].Hash != txOutputs[j].TxHash || o.SlpOutputs[i].Index != txOutputs[j].Index {
				continue
			}
			o.SlpOutputs[i].Output = &model.TxOutput{
				Hash:   txOutputs[j].TxHash,
				Index:  txOutputs[j].Index,
				Amount: txOutputs[j].Value,
				Script: txOutputs[j].LockScript,
			}
			allOutputs = append(allOutputs, o.SlpOutputs[i].Output)
			break
		}
	}
	o.Mutex.Unlock()
	if err := ToOutputs(o.Ctx, GetPrefixFields(o.Fields, "output."), allOutputs); err != nil {
		o.AddError(fmt.Errorf("error attaching to outputs for slp outputs; %w", err))
		return
	}
}
