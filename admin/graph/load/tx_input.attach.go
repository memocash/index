package load

import (
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func AttachToInputs(preloads []string, inputs []*model.TxInput) error {
	if jutil.StringsInSlice([]string{"script", "sequence"}, preloads) {
		if err := attachScriptSequenceToInputs(inputs); err != nil {
			return err
		}
	}
	if jutil.StringInSlice("tx", preloads) {
		if err := attachTxsToInputs(preloads, inputs); err != nil {
			return err
		}
	}
	return nil
}

func attachScriptSequenceToInputs(inputs []*model.TxInput) error {
	var outs []memo.Out
	for i := range inputs {
		if len(inputs[i].Script) > 0 && inputs[i].Sequence > 0 {
			continue
		}
		outs = append(outs, memo.Out{
			TxHash: inputs[i].Hash[:],
			Index:  inputs[i].Index,
		})
	}
	if len(outs) == 0 {
		return nil
	}
	txInputs, err := chain.GetTxInputs(outs)
	if err != nil {
		return fmt.Errorf("error getting tx inputs for model tx inputs script sequence; %w", err)
	}
	for i := range inputs {
		for j := range txInputs {
			if inputs[i].Hash != txInputs[j].TxHash || inputs[i].Index != txInputs[j].Index {
				continue
			}
			inputs[i].Script = txInputs[j].UnlockScript
			inputs[i].Sequence = txInputs[j].Sequence
			break
		}
	}
	return nil
}

func attachTxsToInputs(preloads []string, inputs []*model.TxInput) error {
	var txHashes = make([][32]byte, len(inputs))
	for i := range inputs {
		txHashes[i] = inputs[i].Hash
	}
	txs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting txs for model tx inputs; %w", err)
	}
	for i := range inputs {
		for j := range txs {
			if inputs[i].Hash != txs[j].TxHash {
				continue
			}
			inputs[i].Tx = &model.Tx{
				Hash:     txs[j].TxHash,
				Version:  txs[j].Version,
				LockTime: txs[j].LockTime,
			}
			break
		}
	}
	var allTxs []*model.Tx
	for _, input := range inputs {
		if input.Tx == nil {
			continue
		}
		allTxs = append(allTxs, input.Tx)
	}
	if err := AttachToTxs(GetPrefixPreloads(preloads, "tx."), allTxs); err != nil {
		return err
	}
	return nil
}
