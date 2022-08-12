package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

func (r *roomResolver) Posts(ctx context.Context, obj *model.Room, start *int) ([]*model.Post, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *roomResolver) Followers(ctx context.Context, obj *model.Room, start *int) ([]*model.Profile, error) {
	panic(fmt.Errorf("not implemented"))
}

// Room returns generated.RoomResolver implementation.
func (r *Resolver) Room() generated.RoomResolver { return &roomResolver{r} }

type roomResolver struct{ *Resolver }
