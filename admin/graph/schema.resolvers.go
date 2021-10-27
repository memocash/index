package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jchavannes/jgo/jutil"

	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
)

func (r *mutationResolver) Null(ctx context.Context) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Tx(ctx context.Context, hash *string) (*model.Tx, error) {
	var hashString = "123-single"
	if hash != nil {
		hashString += " - " + *hash
	}
	preloads := GetPreloads(ctx)
	fmt.Printf("preloads: %#v\n", preloads)
	var raw string
	if jutil.StringInSlice("raw", preloads) {
		raw = "456"
	} else {
		fmt.Println("raw not requested")
	}
	return &model.Tx{
		Hash: hashString,
		Raw:  raw,
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

func GetPreloads(ctx context.Context) []string {
	return GetNestedPreloads(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

func GetNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) (preloads []string) {
	for _, column := range fields {
		prefixColumn := GetPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(preloads, GetNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), prefixColumn)...)
	}
	return
}

func GetPreloadString(prefix, name string) string {
	if len(prefix) > 0 {
		return prefix + "." + name
	}
	return name
}
