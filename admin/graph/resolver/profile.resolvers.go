package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

func (r *followResolver) Tx(ctx context.Context, obj *model.Follow) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *followResolver) Lock(ctx context.Context, obj *model.Follow) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting lock from loader for follow resolver", err)
	}
	return lock, nil
}

func (r *followResolver) FollowLock(ctx context.Context, obj *model.Follow) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.FollowLockHash)
	if err != nil {
		return nil, jerr.Get("error getting follow lock from loader for follow resolver", err)
	}
	return lock, nil
}

func (r *postResolver) Tx(ctx context.Context, obj *model.Post) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *postResolver) Lock(ctx context.Context, obj *model.Post) (*model.Lock, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *profileResolver) Lock(ctx context.Context, obj *model.Profile) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting lock from loader for profile resolver", err)
	}
	return lock, nil
}

func (r *profileResolver) Following(ctx context.Context, obj *model.Profile, start *int) ([]*model.Follow, error) {
	address, err := dataloader.NewLockAddressLoader(lockAddressLoaderConfig).Load(obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting address from lock dataloader for profile resolver", err)
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
			TxHash:         hs.GetTxString(memoFollow.TxHash),
			LockHash:       hex.EncodeToString(memoFollow.LockHash),
			FollowLockHash: hex.EncodeToString(memoFollow.Follow),
			Unfollow:       memoFollow.Unfollow,
		})
	}
	return follows, nil
}

func (r *profileResolver) Followers(ctx context.Context, obj *model.Profile, start *int) ([]*model.Follow, error) {
	address, err := dataloader.NewLockAddressLoader(lockAddressLoaderConfig).Load(obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting address from lock dataloader for profile resolver", err)
	}
	var startInt int64
	if start != nil {
		startInt = int64(*start)
	}
	memoFolloweds, err := item.GetMemoFolloweds(ctx, script.GetLockHashForAddress(*address), startInt)
	if err != nil {
		return nil, jerr.Get("error getting memo follows for address", err)
	}
	var follows []*model.Follow
	for _, memoFollowed := range memoFolloweds {
		follows = append(follows, &model.Follow{
			TxHash:         hs.GetTxString(memoFollowed.TxHash),
			LockHash:       hex.EncodeToString(memoFollowed.LockHash),
			FollowLockHash: hex.EncodeToString(memoFollowed.FollowLockHash),
			Unfollow:       memoFollowed.Unfollow,
		})
	}
	return follows, nil
}

func (r *profileResolver) Posts(ctx context.Context, obj *model.Profile, start *int) ([]*model.Post, error) {
	lockHash, err := hex.DecodeString(obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error decoding lock hash for profile resolver", err)
	}
	memoPosts, err := item.GetMemoPost(ctx, [][]byte{lockHash})
	if err != nil {
		return nil, jerr.Get("error getting memo posts for profile resolver", err)
	}
	var posts = make([]*model.Post, len(memoPosts))
	for i, memoPost := range memoPosts {
		posts[i] = &model.Post{
			TxHash:   hs.GetTxString(memoPost.TxHash),
			LockHash: hex.EncodeToString(memoPost.LockHash),
			Text:     memoPost.Post,
		}
	}
	return posts, nil
}

func (r *setNameResolver) Tx(ctx context.Context, obj *model.SetName) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *setNameResolver) Lock(ctx context.Context, obj *model.SetName) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting lock from loader for set name resolver", err)
	}
	return lock, nil
}

func (r *setPicResolver) Tx(ctx context.Context, obj *model.SetPic) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *setPicResolver) Lock(ctx context.Context, obj *model.SetPic) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting lock from loader for set pic resolver", err)
	}
	return lock, nil
}

func (r *setProfileResolver) Tx(ctx context.Context, obj *model.SetProfile) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *setProfileResolver) Lock(ctx context.Context, obj *model.SetProfile) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Get("error getting lock from loader for set profile resolver", err)
	}
	return lock, nil
}

// Follow returns generated.FollowResolver implementation.
func (r *Resolver) Follow() generated.FollowResolver { return &followResolver{r} }

// Post returns generated.PostResolver implementation.
func (r *Resolver) Post() generated.PostResolver { return &postResolver{r} }

// Profile returns generated.ProfileResolver implementation.
func (r *Resolver) Profile() generated.ProfileResolver { return &profileResolver{r} }

// SetName returns generated.SetNameResolver implementation.
func (r *Resolver) SetName() generated.SetNameResolver { return &setNameResolver{r} }

// SetPic returns generated.SetPicResolver implementation.
func (r *Resolver) SetPic() generated.SetPicResolver { return &setPicResolver{r} }

// SetProfile returns generated.SetProfileResolver implementation.
func (r *Resolver) SetProfile() generated.SetProfileResolver { return &setProfileResolver{r} }

type followResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type profileResolver struct{ *Resolver }
type setNameResolver struct{ *Resolver }
type setPicResolver struct{ *Resolver }
type setProfileResolver struct{ *Resolver }
