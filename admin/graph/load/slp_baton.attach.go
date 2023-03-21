package load

import (
	"encoding/hex"
	"fmt"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/slp"
)

type SlpBatons struct {
	baseA
	SlpBatons []*model.SlpBaton
}

func AttachToSlpBatons(preloads []string, slpBatons []*model.SlpBaton) error {
	if len(slpBatons) == 0 {
		return nil
	}
	o := SlpBatons{
		baseA:     baseA{Preloads: preloads},
		SlpBatons: slpBatons,
	}
	o.Wait.Add(1)
	go o.AttachGeneses()
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

func (o *SlpBatons) AttachGeneses() {
	defer o.Wait.Done()
	if !o.HasPreload([]string{"genesis"}) {
		return
	}
	slpGeneses, err := slp.GetGeneses(o.GetTokenHashes())
	if err != nil {
		o.AddError(fmt.Errorf("error getting slp geneses from dataloader; %w", err))
		return
	}
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
			break
		}
	}
	o.Mutex.Unlock()
}
