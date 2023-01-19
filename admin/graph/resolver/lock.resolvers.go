package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

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
func (r *lockResolver) Profile(ctx context.Context, obj *model.Lock) (*model.Profile, error) {
	profile, err := dataloader.NewProfileLoader(load.ProfileLoaderConfig).Load(obj.Address)
	if err != nil {
		return nil, jerr.Get("error getting profile from dataloader for lock resolver", err)
	}
	return profile, nil
}

// Txs is the resolver for the txs field.
func (r *lockResolver) Txs(ctx context.Context, obj *model.Lock, start *model.Date, tx *string) ([]*model.Tx, error) {
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for lock txs resolver", err)
	}
	var startUid []byte
	if start != nil {
		startUid = jutil.CombineBytes(
			address[:],
			jutil.GetTimeByte(time.Time(*start)),
		)
		if tx != nil {
			txHash, err := chainhash.NewHashFromStr(*tx)
			if err != nil {
				return nil, jerr.Get("error decoding start hash for lock txs resolver", err)
			}
			startUid = append(startUid, jutil.ByteReverse(txHash[:])...)
		}
	}
	seenTxs, err := addr.GetSeenTxs(*address, startUid)
	if err != nil {
		return nil, jerr.Get("error getting addr seen txs for lock txs resolver", err)
	}
	var txHashes = make([]string, len(seenTxs))
	for i := range seenTxs {
		txHashes[i] = chainhash.Hash(seenTxs[i].TxHash).String()
	}
	modelTxs, err := TxsLoader(ctx, txHashes)
	if err != nil {
		return nil, jerr.Get("error getting model txs for lock txs resolver", err)
	}
	return modelTxs, nil
}

// Lock returns generated.LockResolver implementation.
func (r *Resolver) Lock() generated.LockResolver { return &lockResolver{r} }

type lockResolver struct{ *Resolver }
