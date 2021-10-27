package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *mutationResolver) Null(ctx context.Context) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Tx(ctx context.Context) (*model.Tx, error) {
	return &model.Tx{
		Hash: "123-single",
		Raw:  "456",
	}, nil
}

func (r *queryResolver) Txs(ctx context.Context) ([]*model.Tx, error) {
	return []*model.Tx{{
		Hash: "123",
		Raw:  "456",
	}}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
