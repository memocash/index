package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *txInputResolver) Tx(ctx context.Context, obj *model.TxInput) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *txInputResolver) Output(ctx context.Context, obj *model.TxInput) (*model.TxOutput, error) {
	return &model.TxOutput{
		Hash:  obj.PrevHash,
		Index: obj.PrevIndex,
	}, nil
}

func (r *txInputResolver) DoubleSpend(ctx context.Context, obj *model.TxInput) (*model.DoubleSpend, error) {
	panic(fmt.Errorf("not implemented"))
}

// TxInput returns generated.TxInputResolver implementation.
func (r *Resolver) TxInput() generated.TxInputResolver { return &txInputResolver{r} }

type txInputResolver struct{ *Resolver }
