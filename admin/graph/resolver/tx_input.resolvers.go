package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

// Tx is the resolver for the tx field.
func (r *txInputResolver) Tx(ctx context.Context, obj *model.TxInput) (*model.Tx, error) {
	var tx = &model.Tx{
		Hash: obj.Hash,
	}
	if HasFieldAny(ctx, []string{"raw"}) {
		txRaw, err := dataloader.NewTxRawLoader(txRawLoaderConfig).Load(obj.Hash)
		if err != nil {
			return nil, jerr.Get("error getting tx raw for output from loader", err)
		}
		tx.Raw = txRaw
	}
	return tx, nil
}

// Output is the resolver for the output field.
func (r *txInputResolver) Output(ctx context.Context, obj *model.TxInput) (*model.TxOutput, error) {
	txOutputs, err := dataloader.NewTxOutputLoader(txOutputLoaderConfig).Load(model.HashIndex{
		Hash:  obj.PrevHash,
		Index: obj.PrevIndex,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx outputs for spends from loader", err)
	}
	return txOutputs, nil
}

// DoubleSpend is the resolver for the double_spend field.
func (r *txInputResolver) DoubleSpend(ctx context.Context, obj *model.TxInput) (*model.DoubleSpend, error) {
	panic(fmt.Errorf("not implemented"))
}

// TxInput returns generated.TxInputResolver implementation.
func (r *Resolver) TxInput() generated.TxInputResolver { return &txInputResolver{r} }

type txInputResolver struct{ *Resolver }
