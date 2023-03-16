package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Tx is the resolver for the tx field.
func (r *txOutputResolver) Tx(ctx context.Context, obj *model.TxOutput) (*model.Tx, error) {
	var tx = &model.Tx{Hash: obj.Hash}
	if err := load.AttachAllToTxs(load.GetPreloads(ctx), []*model.Tx{tx}); err != nil {
		return nil, jerr.Get("error attaching all to output tx", err)
	}
	return tx, nil
}

// Spends is the resolver for the spends field.
func (r *txOutputResolver) Spends(ctx context.Context, obj *model.TxOutput) ([]*model.TxInput, error) {
	var txOutputSpendDataLoader *dataloader.TxOutputSpendLoader
	if load.HasField(ctx, "script") {
		txOutputSpendDataLoader = load.TxOutputSpendWithScript
	} else {
		txOutputSpendDataLoader = load.TxOutputSpend
	}
	txInputs, err := txOutputSpendDataLoader.Load(model.HashIndex{
		Hash:  chainhash.Hash(obj.Hash).String(),
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for spends from loader", err)
	}
	return txInputs, nil
}

// Slp is the resolver for the slp field.
func (r *txOutputResolver) Slp(ctx context.Context, obj *model.TxOutput) (*model.SlpOutput, error) {
	slpOutput, err := load.SlpOutput.Load(model.HashIndex{
		Hash:  chainhash.Hash(obj.Hash).String(),
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting slp output for tx output from loader", err)
	}
	return slpOutput, nil
}

// SlpBaton is the resolver for the slp_baton field.
func (r *txOutputResolver) SlpBaton(ctx context.Context, obj *model.TxOutput) (*model.SlpBaton, error) {
	slpBaton, err := load.SlpBaton.Load(model.HashIndex{
		Hash:  chainhash.Hash(obj.Hash).String(),
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting slp baton for tx output from loader", err)
	}
	return slpBaton, nil
}

// Lock is the resolver for the lock field.
func (r *txOutputResolver) Lock(ctx context.Context, obj *model.TxOutput) (*model.Lock, error) {
	if len(obj.Script) == 0 {
		return nil, nil
	}
	var modelLock = &model.Lock{
		Address: wallet.GetAddressStringFromPkScript(obj.Script),
	}
	if load.HasField(ctx, "balance") {
		// TODO: Reimplement if needed
		return nil, jerr.New("error balance no longer implemented")
	}
	return modelLock, nil
}

// TxOutput returns generated.TxOutputResolver implementation.
func (r *Resolver) TxOutput() generated.TxOutputResolver { return &txOutputResolver{r} }

type txOutputResolver struct{ *Resolver }
