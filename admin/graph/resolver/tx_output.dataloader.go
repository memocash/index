package resolver

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

var txOutputSpendLoaderConfig = dataloader.TxOutputSpendLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []model.HashIndex) ([][]*model.TxInput, []error) {
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
		outputInputs, err := item.GetOutputInputs(outs)
		if err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting output spends for tx", err)}
		}
		var spends = make([][]*model.TxInput, len(outs))
		for i := range outs {
			for _, outputInput := range outputInputs {
				if bytes.Equal(outs[i].TxHash, outputInput.PrevHash) && outs[i].Index == outputInput.PrevIndex {
					outputInputHash, err := chainhash.NewHash(outputInput.Hash)
					if err != nil {
						return nil, []error{jerr.Get("error getting output spend hash", err)}
					}
					spends[i] = append(spends[i], &model.TxInput{
						Hash:      outputInputHash.String(),
						Index:     outputInput.Index,
						PrevHash:  keys[i].Hash,
						PrevIndex: outputInput.PrevIndex,
					})
				}
			}
		}
		return spends, nil
	},
}
