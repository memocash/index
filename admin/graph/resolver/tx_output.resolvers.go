package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *txOutputResolver) Tx(ctx context.Context, obj *model.TxOutput) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *txOutputResolver) Spends(ctx context.Context, obj *model.TxOutput) ([]*model.TxInput, error) {
	hash, err := chainhash.NewHashFromStr(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error parsing spend tx hash for output", err)
	}
	outputInputs, err := item.GetOutputInput(memo.Out{
		TxHash: hash.CloneBytes(),
		Index:  obj.Index,
	})
	if err != nil && !client.IsResourceUnavailableError(err) {
		return nil, jerr.Get("error getting output spends for tx", err)
	}
	var spends []*model.TxInput
	for _, outputInput := range outputInputs {
		outputInputHash, err := chainhash.NewHash(outputInput.Hash)
		if err != nil {
			return nil, jerr.Get("error getting output spend hash", err)
		}
		spends = append(spends, &model.TxInput{
			Hash:      outputInputHash.String(),
			Index:     outputInput.Index,
			PrevHash:  obj.Hash,
			PrevIndex: outputInput.PrevIndex,
		})
	}
	return spends, nil
}

// TxOutput returns generated.TxOutputResolver implementation.
func (r *Resolver) TxOutput() generated.TxOutputResolver { return &txOutputResolver{r} }

type txOutputResolver struct{ *Resolver }
