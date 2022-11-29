package resolver

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

var txOutputLoaderConfig = dataloader.TxOutputLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []model.HashIndex) ([]*model.TxOutput, []error) {
		var memoOuts = make([]memo.Out, len(keys))
		for i := range keys {
			txHash, err := chainhash.NewHashFromStr(keys[i].Hash)
			if err != nil {
				return nil, []error{jerr.Get("error getting tx hash for inputs", err)}
			}
			memoOuts[i] = memo.Out{
				TxHash: txHash[:],
				Index:  keys[i].Index,
			}
		}
		txOutputs, err := chain.GetTxOutputs(memoOuts)
		if err != nil {
			return nil, []error{jerr.Get("error getting tx outputs for model tx", err)}
		}
		var modelOutputs = make([]*model.TxOutput, len(txOutputs))
		for i := range memoOuts {
			for _, txOutput := range txOutputs {
				if bytes.Equal(memoOuts[i].TxHash, txOutput.TxHash[:]) && memoOuts[i].Index == txOutput.Index {
					modelOutputs[i] = &model.TxOutput{
						Hash:   chainhash.Hash(txOutput.TxHash).String(),
						Index:  txOutput.Index,
						Script: hex.EncodeToString(txOutput.LockScript),
						Amount: txOutput.Value,
					}
					break
				}
			}
		}
		return modelOutputs, nil
	},
}
