package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func (r *followResolver) Tx(ctx context.Context, obj *model.Follow) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *profileResolver) Following(ctx context.Context, obj *model.Profile, start *int) ([]*model.Follow, error) {
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
	var follows []*model.Follow
	for _, memoFollow := range memoFollows {
		follows = append(follows, &model.Follow{
			TxHash:     hs.GetTxString(memoFollow.TxHash),
			Lock:       &model.Lock{Hash: hex.EncodeToString(memoFollow.LockHash)},
			FollowLock: &model.Lock{Hash: hex.EncodeToString(memoFollow.Follow)},
			Unfollow:   memoFollow.Unfollow,
		})
	}
	return follows, nil
}

func (r *profileResolver) Followers(ctx context.Context, obj *model.Profile, start *int) ([]*model.Follow, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *setNameResolver) Tx(ctx context.Context, obj *model.SetName) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *setPicResolver) Tx(ctx context.Context, obj *model.SetPic) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *setProfileResolver) Tx(ctx context.Context, obj *model.SetProfile) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// Follow returns generated.FollowResolver implementation.
func (r *Resolver) Follow() generated.FollowResolver { return &followResolver{r} }

// Profile returns generated.ProfileResolver implementation.
func (r *Resolver) Profile() generated.ProfileResolver { return &profileResolver{r} }

// SetName returns generated.SetNameResolver implementation.
func (r *Resolver) SetName() generated.SetNameResolver { return &setNameResolver{r} }

// SetPic returns generated.SetPicResolver implementation.
func (r *Resolver) SetPic() generated.SetPicResolver { return &setPicResolver{r} }

// SetProfile returns generated.SetProfileResolver implementation.
func (r *Resolver) SetProfile() generated.SetProfileResolver { return &setProfileResolver{r} }

type followResolver struct{ *Resolver }
type profileResolver struct{ *Resolver }
type setNameResolver struct{ *Resolver }
type setPicResolver struct{ *Resolver }
type setProfileResolver struct{ *Resolver }
