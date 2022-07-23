package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

func (r *profileResolver) Following(ctx context.Context, obj *model.Profile, start *int) ([]*model.Profile, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *profileResolver) Followers(ctx context.Context, obj *model.Profile, start *int) ([]*model.Profile, error) {
	panic(fmt.Errorf("not implemented"))
}

// Profile returns generated.ProfileResolver implementation.
func (r *Resolver) Profile() generated.ProfileResolver { return &profileResolver{r} }

type profileResolver struct{ *Resolver }
