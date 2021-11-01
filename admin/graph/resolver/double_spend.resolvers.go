package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *doubleSpendResolver) Output(ctx context.Context, obj *model.DoubleSpend) (*model.TxOutput, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *doubleSpendResolver) Inputs(ctx context.Context, obj *model.DoubleSpend) ([]*model.TxInput, error) {
	panic(fmt.Errorf("not implemented"))
}

// DoubleSpend returns generated.DoubleSpendResolver implementation.
func (r *Resolver) DoubleSpend() generated.DoubleSpendResolver { return &doubleSpendResolver{r} }

type doubleSpendResolver struct{ *Resolver }
