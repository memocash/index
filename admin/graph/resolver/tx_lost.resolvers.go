package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

func (r *txLostResolver) Tx(ctx context.Context, obj *model.TxLost) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// TxLost returns generated.TxLostResolver implementation.
func (r *Resolver) TxLost() generated.TxLostResolver { return &txLostResolver{r} }

type txLostResolver struct{ *Resolver }
