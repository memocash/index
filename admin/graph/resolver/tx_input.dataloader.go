package resolver

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/dataloader"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"
	"time"
)

var txInputOutputLoaderConfig = dataloader.TxOutputLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []model.HashIndex) ([]*model.TxOutput, []error) {
		var outs = make([]memo.Out, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i].Hash)
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			outs[i] = memo.Out{
				TxHash: hash.CloneBytes(),
				Index:  keys[i].Index,
			}
		}
		txOutputs, err := item.GetTxOutputs(outs)
		if err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting outputs for tx inputs", err)}
		}
		var outputs = make([]*model.TxOutput, len(outs))
		for i := range outs {
			for _, txOutput := range txOutputs {
				if bytes.Equal(txOutput.TxHash, outs[i].TxHash) && txOutput.Index == outs[i].Index {
					outputHash, err := chainhash.NewHash(txOutput.TxHash)
					if err != nil {
						return nil, []error{jerr.Get("error getting input output hash", err)}
					}
					outputs[i] = &model.TxOutput{
						Hash:   outputHash.String(),
						Index:  txOutput.Index,
						Amount: txOutput.Value,
					}
					break
				}
			}
		}
		return outputs, nil
	},
}
