package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/dataloader"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *txOutputResolver) Tx(ctx context.Context, obj *model.TxOutput) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *txOutputResolver) Spends(ctx context.Context, obj *model.TxOutput) ([]*model.TxInput, error) {
	txInputs, err := dataloader.NewTxOutputSpendLoader(txOutputSpendLoaderConfig).Load(model.HashIndex{
		Hash:  obj.Hash,
		Index: obj.Index,
	})
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for spends from loader", err)
	}
	return txInputs, nil
}

func (r *txOutputResolver) DoubleSpend(ctx context.Context, obj *model.TxOutput) (*model.DoubleSpend, error) {
	panic(fmt.Errorf("not implemented"))
}

// TxOutput returns generated.TxOutputResolver implementation.
func (r *Resolver) TxOutput() generated.TxOutputResolver { return &txOutputResolver{r} }

type txOutputResolver struct{ *Resolver }
