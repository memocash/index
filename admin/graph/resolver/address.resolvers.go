package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Profile is the resolver for the profile field.
func (r *addrResolver) Profile(ctx context.Context, obj *model.Addr) (*model.Profile, error) {
	profile, err := dataloader.NewProfileLoader(load.ProfileLoaderConfig).Load(obj.Address)
	if err != nil {
		return nil, jerr.Get("error getting profile from dataloader for lock resolver", err)
	}
	return profile, nil
}

// Utxos is the resolver for the utxos field.
func (r *addrResolver) Utxos(ctx context.Context, obj *model.Addr, start *model.HashIndex) ([]*model.TxOutput, error) {
	// TODO: Fix UTXO resolver
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Get("error getting address for utxo resolver", err)
	}
	var startUid []byte
	if start != nil {
		startHash, err := chainhash.NewHashFromStr(start.Hash)
		if err != nil {
			return nil, jerr.Get("error decoding start hash", err)
		}
		// TODO: Support height
		startUid = addr.GetHeightTxHashIndexUid(*address, 0, *startHash, start.Index)
	}
	heightOutputs, err := addr.GetHeightOutputs(*address, startUid)
	if err != nil {
		return nil, jerr.Get("error getting height outputs for addr utxo resolver", err)
	}
	var txOutputs = make([]*model.TxOutput, len(heightOutputs))
	for i := range heightOutputs {
		txOutputs[i] = &model.TxOutput{
			Hash:   chainhash.Hash(heightOutputs[i].TxHash).String(),
			Index:  heightOutputs[i].Index,
			Amount: heightOutputs[i].Value,
		}
	}
	return txOutputs, nil
}

// Outputs is the resolver for the outputs field.
func (r *addrResolver) Outputs(ctx context.Context, obj *model.Addr, start *model.HashIndex, height *int) ([]*model.TxOutput, error) {
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

// Addr returns generated.AddrResolver implementation.
func (r *Resolver) Addr() generated.AddrResolver { return &addrResolver{r} }

type addrResolver struct{ *Resolver }
