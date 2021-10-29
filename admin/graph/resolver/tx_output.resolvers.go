package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/dataloader"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"
	"time"
)

func (r *txOutputResolver) Tx(ctx context.Context, obj *model.TxOutput) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *txOutputResolver) Spends(ctx context.Context, obj *model.TxOutput) ([]*model.TxInput, error) {
	txOutputSpendLoader := dataloader.NewTxOutputSpendLoader(dataloader.TxOutputSpendLoaderConfig{
		Wait:     2 * time.Millisecond,
		MaxBatch: 100,
		Fetch: func(keys []model.HashIndex) ([][]*model.TxInput, []error) {
			var outs = make([]memo.Out, len(keys))
			for i := range keys {
				hash, err := chainhash.NewHashFromStr(obj.Hash)
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
							PrevHash:  obj.Hash,
							PrevIndex: outputInput.PrevIndex,
						})
					}
				}
			}
			return spends, nil
		},
	})
	txInputs, err := txOutputSpendLoader.Load(model.HashIndex{
		Hash:  obj.Hash,
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for spends from loader", err)
	}
	return txInputs, nil
}

// TxOutput returns generated.TxOutputResolver implementation.
func (r *Resolver) TxOutput() generated.TxOutputResolver { return &txOutputResolver{r} }

type txOutputResolver struct{ *Resolver }
