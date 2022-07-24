package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"

	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

func (r *lockResolver) Profile(ctx context.Context, obj *model.Lock) (*model.Profile, error) {
	profile, err := dataloader.NewProfileLoader(profileLoaderConfig).Load(obj.Address)
	if err != nil {
		return nil, jerr.Get("error getting profile from dataloader for lock resolver", err)
	}
	return profile, nil
}

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

func (r *lockResolver) Outputs(ctx context.Context, obj *model.Lock, start *model.HashIndex, height *int) ([]*model.TxOutput, error) {
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
		var height64 int64
		if height != nil {
			height64 = int64(*height)
		}
		startUid = item.GetLockHeightOutputUid(lockHash, height64, startHash, start.Index)
	} else if height != nil {
		startUid = jutil.CombineBytes(lockHash, jutil.GetInt64DataBig(int64(*height)))
	}
	lockHeightOutputs, err := item.GetLockHeightOutputs(lockHash, startUid)
	if err != nil {
		return nil, jerr.Get("error getting lock outputs for lock output resolver", err)
	}
	var outs = make([]memo.Out, len(lockHeightOutputs))
	for i := range lockHeightOutputs {
		outs[i] = memo.Out{
			TxHash: lockHeightOutputs[i].Hash,
			Index:  lockHeightOutputs[i].Index,
		}
	}
	txOutputs, err := item.GetTxOutputs(outs)
	if err != nil {
		return nil, jerr.Get("error getting tx outputs for lock resolver", err)
	}
	var modelTxOutputs = make([]*model.TxOutput, len(lockHeightOutputs))
	for i := range txOutputs {
		modelTxOutputs[i] = &model.TxOutput{
			Hash:   hs.GetTxString(txOutputs[i].TxHash),
			Index:  txOutputs[i].Index,
			Amount: txOutputs[i].Value,
		}
	}
	return modelTxOutputs, nil
}

// Lock returns generated.LockResolver implementation.
func (r *Resolver) Lock() generated.LockResolver { return &lockResolver{r} }

type lockResolver struct{ *Resolver }
