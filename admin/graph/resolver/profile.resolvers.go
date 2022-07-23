package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"

	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

func (r *profileResolver) Following(ctx context.Context, obj *model.Profile, start *int) ([]*model.Profile, error) {
	address, err := wallet.GetAddressFromStringErr(obj.Lock.Address)
	if err != nil {
		return nil, jerr.Get("error getting address from string", err)
	}
	var startInt int64
	if start != nil {
		startInt = int64(*start)
	}
	memoFollows, err := item.GetMemoFollows(ctx, script.GetLockHashForAddress(*address), startInt)
	if err != nil {
		return nil, jerr.Get("error getting memo follows for address", err)
	}
	var profiles []*model.Profile
	for _, memoFollow := range memoFollows {
		profiles = append(profiles, &model.Profile{
			Lock: &model.Lock{
				Hash: hex.EncodeToString(memoFollow.LockHash),
			},
		})
	}
	return profiles, nil
}

func (r *profileResolver) Followers(ctx context.Context, obj *model.Profile, start *int) ([]*model.Profile, error) {
	panic(fmt.Errorf("not implemented"))
}

// Profile returns generated.ProfileResolver implementation.
func (r *Resolver) Profile() generated.ProfileResolver { return &profileResolver{r} }

type profileResolver struct{ *Resolver }
