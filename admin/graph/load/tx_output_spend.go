package load

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func txOutputSpend(keys []model.HashIndex, withScript bool) ([][]*model.TxInput, []error) {
	var outs = make([]memo.Out, len(keys))
	for i := range keys {
		hash, err := chainhash.NewHashFromStr(keys[i].Hash)
		if err != nil {
			return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
		}
		outs[i] = memo.Out{
			TxHash: hash[:],
			Index:  keys[i].Index,
		}
	}
	outputInputs, err := chain.GetOutputInputs(outs)
	if err != nil && !client.IsResourceUnavailableError(err) {
		return nil, []error{jerr.Get("error getting output spends for tx", err)}
	}
	var txInputs []*chain.TxInput
	if withScript {
		var ins = make([]memo.Out, len(outputInputs))
		for i := range outputInputs {
			ins[i] = memo.Out{
				TxHash: outputInputs[i].Hash[:],
				Index:  outputInputs[i].Index,
			}
		}
		if txInputs, err = chain.GetTxInputs(ins); err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting tx inputs for tx", err)}
		}
	}
	var spends = make([][]*model.TxInput, len(outs))
	for i := range outs {
		for _, outputInput := range outputInputs {
			if bytes.Equal(outs[i].TxHash, outputInput.PrevHash[:]) && outs[i].Index == outputInput.PrevIndex {
				var modelTxInput = &model.TxInput{
					Hash:      chainhash.Hash(outputInput.Hash).String(),
					Index:     outputInput.Index,
					PrevHash:  keys[i].Hash,
					PrevIndex: outputInput.PrevIndex,
				}
				for _, txInput := range txInputs {
					if txInput.TxHash == outputInput.Hash && txInput.Index == outputInput.Index {
						modelTxInput.Script = hex.EncodeToString(txInput.UnlockScript)
						break
					}
				}
				spends[i] = append(spends[i], modelTxInput)
				break
			}
		}
	}
	return spends, nil
}

var TxOutputSpend = dataloader.NewTxOutputSpendLoader(dataloader.TxOutputSpendLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []model.HashIndex) ([][]*model.TxInput, []error) {
		return txOutputSpend(keys, false)
	},
})

var TxOutputSpendWithScript = dataloader.NewTxOutputSpendLoader(dataloader.TxOutputSpendLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []model.HashIndex) ([][]*model.TxInput, []error) {
		return txOutputSpend(keys, true)
	},
})
