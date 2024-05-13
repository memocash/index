package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Txs is the resolver for the txs field.
func (r *lockResolver) Txs(ctx context.Context, obj *model.Lock, start *model.Date, tx *string) ([]*model.Tx, error) {
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for lock txs resolver", err)
	}
	var startUid []byte
	if start != nil && time.Time(*start).After(memo.GetGenesisTime()) {
		startUid = jutil.CombineBytes(
			address[:],
			jutil.GetTimeByteNanoBig(time.Time(*start)),
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
	var modelTxs = make([]*model.Tx, len(seenTxs))
	for i := range seenTxs {
		modelTxs[i] = &model.Tx{
			Hash: seenTxs[i].TxHash,
			Seen: model.Date(seenTxs[i].Seen),
		}
	}
	if err := load.AttachToTxs(ctx, load.GetFields(ctx), modelTxs); err != nil {
		return nil, jerr.Get("error attaching all to txs for lock txs resolver", err)
	}
	return modelTxs, nil
}

// Lock returns generated.LockResolver implementation.
func (r *Resolver) Lock() generated.LockResolver { return &lockResolver{r} }

type lockResolver struct{ *Resolver }
