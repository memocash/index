package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
)

// Tx is the resolver for the tx field.
func (r *txInputResolver) Tx(ctx context.Context, obj *model.TxInput) (*model.Tx, error) {
	var tx = &model.Tx{Hash: obj.Hash}
	if err := load.AttachAllToTxs(load.GetPreloads(ctx), []*model.Tx{tx}); err != nil {
		return nil, jerr.Get("error attaching all to input tx", err)
	}
	return tx, nil
}

// Output is the resolver for the output field.
func (r *txInputResolver) Output(ctx context.Context, obj *model.TxInput) (*model.TxOutput, error) {
	txOutput, err := load.TxOutput.Load(model.HashIndex{
		Hash:  chainhash.Hash(obj.PrevHash).String(),
		Index: obj.PrevIndex,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx output for spends from loader", err)
	}
	return txOutput, nil
}

// TxInput returns generated.TxInputResolver implementation.
func (r *Resolver) TxInput() generated.TxInputResolver { return &txInputResolver{r} }

type txInputResolver struct{ *Resolver }
