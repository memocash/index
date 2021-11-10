package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/item"
)

func (r *lockResolver) Utxos(ctx context.Context, obj *model.Lock) ([]*model.TxOutput, error) {
	hash, err := hex.DecodeString(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for utxo resolver", err)
	}
	lockUtxos, err := item.GetLockUtxos(hash, nil)
	if err != nil {
		return nil, jerr.Get("error getting lock utxos for lock utxo resolver", err)
	}
	var txOutputs = make([]*model.TxOutput, len(lockUtxos))
	for i := range lockUtxos {
		txOutputs[i] = &model.TxOutput{
			Hash:   hex.EncodeToString(lockUtxos[i].Hash),
			Index:  lockUtxos[i].Index,
			Amount: lockUtxos[i].Value,
		}
	}
	return txOutputs, nil
}

// Lock returns generated.LockResolver implementation.
func (r *Resolver) Lock() generated.LockResolver { return &lockResolver{r} }

type lockResolver struct{ *Resolver }
