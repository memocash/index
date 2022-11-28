package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Profile is the resolver for the profile field.
func (r *lockResolver) Profile(ctx context.Context, obj *model.Lock) (*model.Profile, error) {
	profile, err := dataloader.NewProfileLoader(load.ProfileLoaderConfig).Load(obj.Address)
	if err != nil {
		return nil, jerr.Get("error getting profile from dataloader for lock resolver", err)
	}
	return profile, nil
}

// Spends is the resolver for the spends field.
func (r *lockResolver) Spends(ctx context.Context, obj *model.Lock, start *model.HashIndex, height *int) ([]*model.TxInput, error) {
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for lock spends resolver", err)
	}
	var startUid []byte
	if start != nil {
		startHash, err := chainhash.NewHashFromStr(start.Hash)
		if err != nil {
			return nil, jerr.Get("error decoding start hash for lock spends resolver", err)
		}
		var height64 int64
		if height != nil {
			height64 = int64(*height)
		}
		startUid = addr.GetHeightTxHashIndexUid(*address, int32(height64), *startHash, start.Index)
	} else if height != nil {
		startUid = jutil.CombineBytes(address[:], jutil.GetInt64DataBig(int64(*height)))
	}
	heightInputs, err := addr.GetHeightInputs(*address, startUid)
	if err != nil {
		return nil, jerr.Get("error getting addr inputs for lock spends resolver", err)
	}
	var outs = make([]memo.Out, len(heightInputs))
	for i := range heightInputs {
		outs[i] = memo.Out{
			TxHash: heightInputs[i].TxHash[:],
			Index:  heightInputs[i].Index,
		}
	}
	txInputs, err := chain.GetTxInputs(outs)
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for lock spends resolver", err)
	}
	var modelTxOutputs = make([]*model.TxInput, len(heightInputs))
	for i := range txInputs {
		modelTxOutputs[i] = &model.TxInput{
			Hash:      chainhash.Hash(txInputs[i].TxHash).String(),
			Index:     txInputs[i].Index,
			PrevHash:  chainhash.Hash(txInputs[i].PrevHash).String(),
			PrevIndex: txInputs[i].PrevIndex,
			Script:    hex.EncodeToString(txInputs[i].UnlockScript),
		}
	}
	return modelTxOutputs, nil
}

// Outputs is the resolver for the outputs field.
func (r *lockResolver) Outputs(ctx context.Context, obj *model.Lock, start *model.HashIndex, height *int) ([]*model.TxOutput, error) {
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for lock output resolver", err)
	}
	var startUid []byte
	if start != nil {
		startHash, err := chainhash.NewHashFromStr(start.Hash)
		if err != nil {
			return nil, jerr.Get("error decoding start hash", err)
		}
		var height64 int64
		if height != nil {
			height64 = int64(*height)
		}
		startUid = addr.GetHeightTxHashIndexUid(*address, int32(height64), *startHash, start.Index)
	} else if height != nil {
		startUid = jutil.CombineBytes(address[:], jutil.GetInt64DataBig(int64(*height)))
	}
	heightOutputs, err := addr.GetHeightOutputs(*address, startUid)
	if err != nil {
		return nil, jerr.Get("error getting addr outputs for addr output resolver", err)
	}
	var modelTxOutputs = make([]*model.TxOutput, len(heightOutputs))
	for i := range heightOutputs {
		modelTxOutputs[i] = &model.TxOutput{
			Hash:   chainhash.Hash(heightOutputs[i].TxHash).String(),
			Index:  heightOutputs[i].Index,
			Amount: heightOutputs[i].Value,
		}
	}
	return modelTxOutputs, nil
}

// Lock returns generated.LockResolver implementation.
func (r *Resolver) Lock() generated.LockResolver { return &lockResolver{r} }

type lockResolver struct{ *Resolver }
