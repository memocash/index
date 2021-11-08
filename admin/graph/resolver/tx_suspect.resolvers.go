package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *txSuspectResolver) Tx(ctx context.Context, obj *model.TxSuspect) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// TxSuspect returns generated.TxSuspectResolver implementation.
func (r *Resolver) TxSuspect() generated.TxSuspectResolver { return &txSuspectResolver{r} }

type txSuspectResolver struct{ *Resolver }
