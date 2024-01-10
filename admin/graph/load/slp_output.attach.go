package load

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/slp"
)

type SlpOutputs struct {
	baseA
	SlpOutputs []*model.SlpOutput
}

func AttachToSlpOutputs(ctx context.Context, fields []Field, slpOutputs []*model.SlpOutput) error {
	if len(slpOutputs) == 0 {
		return nil
	}
	o := SlpOutputs{
		baseA:      baseA{Ctx: ctx, Fields: fields},
		SlpOutputs: slpOutputs,
	}
	o.Wait.Add(1)
	go o.AttachGeneses()
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

func (o *SlpOutputs) AttachGeneses() {
	defer o.Wait.Done()
	if !o.HasField([]string{"genesis"}) {
		return
	}
	slpGeneses, err := slp.GetGeneses(o.Ctx, o.GetTokenHashes())
	if err != nil {
		o.AddError(fmt.Errorf("error getting slp geneses from dataloader; %w", err))
		return
	}
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
			break
		}
	}
	o.Mutex.Unlock()
}
