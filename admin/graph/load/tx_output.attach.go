package load

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func AttachToOutputs(preloads []string, outputs []*model.TxOutput) error {
	if jutil.StringsInSlice([]string{"amount", "script"}, preloads) {
		if err := attachInfoToOutputs(outputs); err != nil {
			return err
		}
	}
	if jutil.StringInSlice("spends", preloads) {
		if err := attachSpendsToOutput(preloads, outputs); err != nil {
			return err
		}
	}
	return nil
}

func attachInfoToOutputs(outputs []*model.TxOutput) error {
	var outs []memo.Out
	for i := range outputs {
		if outputs[i].Amount == 0 || len(outputs[i].Script) == 0 {
			outs = append(outs, memo.Out{
				TxHash: outputs[i].Hash[:],
				Index:  outputs[i].Index,
			})
		}
	}
	if len(outs) == 0 {
		return nil
	}
	txOutputs, err := chain.GetTxOutputs(outs)
	if err != nil {
		return fmt.Errorf("error getting tx outputs for model tx; %w", err)
	}
	for i := range outputs {
		for j := range txOutputs {
			if outputs[i].Hash != txOutputs[j].TxHash || outputs[i].Index != txOutputs[j].Index {
				continue
			}
			outputs[i].Amount = txOutputs[j].Value
			outputs[i].Script = txOutputs[j].LockScript
			break
		}
	}
	return nil
}

func attachSpendsToOutput(preloads []string, outputs []*model.TxOutput) error {
	var outs = make([]memo.Out, len(outputs))
	for i := range outputs {
		outs[i] = memo.Out{
			TxHash: outputs[i].Hash[:],
			Index:  outputs[i].Index,
		}
	}
	spends, err := chain.GetOutputInputs(outs)
	if err != nil {
		return fmt.Errorf("error getting tx inputs spends for model tx outputs; %w", err)
	}
	for i := range outputs {
		for j := range spends {
			if outputs[i].Hash != spends[j].PrevHash || outputs[i].Index != spends[j].PrevIndex {
				continue
			}
			outputs[i].Spends = append(outputs[i].Spends, &model.TxInput{
				Hash:      spends[j].Hash,
				Index:     spends[j].Index,
				PrevHash:  spends[j].PrevHash,
				PrevIndex: spends[j].PrevIndex,
			})
		}
	}
	var allSpends []*model.TxInput
	for _, output := range outputs {
		allSpends = append(allSpends, output.Spends...)
	}
	if err := AttachToInputs(GetPrefixPreloads(preloads, "spends."), allSpends); err != nil {
		return err
	}
	return nil
}
