package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *blockResolver) Timestamp(ctx context.Context, obj *model.Block) (*model.Date, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *blockResolver) Txs(ctx context.Context, obj *model.Block) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// Block returns generated.BlockResolver implementation.
func (r *Resolver) Block() generated.BlockResolver { return &blockResolver{r} }

type blockResolver struct{ *Resolver }
