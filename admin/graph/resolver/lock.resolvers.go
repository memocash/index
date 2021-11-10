package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"github.com/memocash/server/ref/bitcoin/tx/hs"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/item"
)

func (r *lockResolver) Utxos(ctx context.Context, obj *model.Lock, start *model.HashIndex) ([]*model.TxOutput, error) {
	lockHash, err := hex.DecodeString(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for utxo resolver", err)
	}
	var startUid []byte
	if start != nil {
		startHash, err := hex.DecodeString(start.Hash)
		if err != nil {
			return nil, jerr.Get("error decoding start hash", err)
		}
		startUid = item.GetLockOutputUid(lockHash, startHash, start.Index)
	}
	lockUtxos, err := item.GetLockUtxos(lockHash, startUid)
	if err != nil {
		return nil, jerr.Get("error getting lock utxos for lock utxo resolver", err)
	}
	var txOutputs = make([]*model.TxOutput, len(lockUtxos))
	for i := range lockUtxos {
		txOutputs[i] = &model.TxOutput{
			Hash:   hs.GetTxString(lockUtxos[i].Hash),
			Index:  lockUtxos[i].Index,
			Amount: lockUtxos[i].Value,
		}
	}
	return txOutputs, nil
}

func (r *lockResolver) Outputs(ctx context.Context, obj *model.Lock, start *model.HashIndex) ([]*model.TxOutput, error) {
	lockHash, err := hex.DecodeString(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for lock output resolver", err)
	}
	var startUid []byte
	if start != nil {
		startHash, err := hex.DecodeString(start.Hash)
		if err != nil {
			return nil, jerr.Get("error decoding start hash", err)
		}
		startUid = item.GetLockOutputUid(lockHash, startHash, start.Index)
	}
	lockOutputs, err := item.GetLockOutputs(lockHash, startUid)
	if err != nil {
		return nil, jerr.Get("error getting lock outputs for lock output resolver", err)
	}
	var txOutputs = make([]*model.TxOutput, len(lockOutputs))
	for i := range lockOutputs {
		txOutputs[i] = &model.TxOutput{
			Hash:  hs.GetTxString(lockOutputs[i].Hash),
			Index: lockOutputs[i].Index,
		}
	}
	return txOutputs, nil
}

// Lock returns generated.LockResolver implementation.
func (r *Resolver) Lock() generated.LockResolver { return &lockResolver{r} }

type lockResolver struct{ *Resolver }
