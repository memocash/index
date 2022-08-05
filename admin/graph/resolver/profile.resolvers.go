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
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

func (r *followResolver) Tx(ctx context.Context, obj *model.Follow) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for follow resolver: %s", obj.TxHash)
	}
	return tx, nil
}

func (r *followResolver) Lock(ctx context.Context, obj *model.Follow) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for follow resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *followResolver) FollowLock(ctx context.Context, obj *model.Follow) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.FollowLockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting follow lock from loader for follow resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *likeResolver) Tx(ctx context.Context, obj *model.Like) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for like resolver: %s", obj.TxHash)
	}
	return tx, nil
}

func (r *likeResolver) Lock(ctx context.Context, obj *model.Like) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for like resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *likeResolver) Post(ctx context.Context, obj *model.Like) (*model.Post, error) {
	post, err := dataloader.NewPostLoader(load.PostLoaderConfig).Load(obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting post from post dataloader for like resolver: %s", obj.LockHash)
	}
	return post, nil
}

func (r *postResolver) Tx(ctx context.Context, obj *model.Post) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for post resolver: %s", obj.TxHash)
	}
	return tx, nil
}

func (r *postResolver) Lock(ctx context.Context, obj *model.Post) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for post resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *postResolver) Likes(ctx context.Context, obj *model.Post) ([]*model.Like, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *profileResolver) Lock(ctx context.Context, obj *model.Profile) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for profile resolver: %s", obj.LockHash)
	}
	return lock, nil
}

func (r *profileResolver) Following(ctx context.Context, obj *model.Profile, start *int) ([]*model.Follow, error) {
	address, err := dataloader.NewLockAddressLoader(lockAddressLoaderConfig).Load(obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting address from lock dataloader for profile following resolver: %s", obj.LockHash)
	}
	var startInt int64
	if start != nil {
		startInt = int64(*start)
	}
	lockMemoFollows, err := item.GetLockMemoFollows(ctx, script.GetLockHashForAddress(*address), startInt)
	if err != nil {
		return nil, jerr.Get("error getting lock memo follows for address", err)
	}
	var follows []*model.Follow
	for _, lockMemoFollow := range lockMemoFollows {
		follows = append(follows, &model.Follow{
			TxHash:         hs.GetTxString(lockMemoFollow.TxHash),
			LockHash:       hex.EncodeToString(lockMemoFollow.LockHash),
			FollowLockHash: hex.EncodeToString(lockMemoFollow.Follow),
			Unfollow:       lockMemoFollow.Unfollow,
		})
	}
	return follows, nil
}

func (r *profileResolver) Followers(ctx context.Context, obj *model.Profile, start *int) ([]*model.Follow, error) {
	address, err := dataloader.NewLockAddressLoader(lockAddressLoaderConfig).Load(obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting address from lock dataloader for profile followers resolver: %s", obj.LockHash)
	}
	var startInt int64
	if start != nil {
		startInt = int64(*start)
	}
	lockMemoFolloweds, err := item.GetLockMemoFolloweds(ctx, script.GetLockHashForAddress(*address), startInt)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock memo follows for address: %s", obj.LockHash)
	}
	var follows []*model.Follow
	for _, lockMemoFollowed := range lockMemoFolloweds {
		follows = append(follows, &model.Follow{
			TxHash:         hs.GetTxString(lockMemoFollowed.TxHash),
			LockHash:       hex.EncodeToString(lockMemoFollowed.LockHash),
			FollowLockHash: hex.EncodeToString(lockMemoFollowed.FollowLockHash),
			Unfollow:       lockMemoFollowed.Unfollow,
		})
	}
	return follows, nil
}

func (r *profileResolver) Posts(ctx context.Context, obj *model.Profile, start *int) ([]*model.Post, error) {
	lockHash, err := hex.DecodeString(obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error decoding lock hash for profile resolver: %s", obj.LockHash)
	}
	lockMemoPosts, err := item.GetLockMemoPosts(ctx, [][]byte{lockHash})
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock memo posts for profile resolver: %s", obj.LockHash)
	}
	var postTxHashes = make([][]byte, len(lockMemoPosts))
	for i := range lockMemoPosts {
		postTxHashes[i] = lockMemoPosts[i].TxHash
	}
	memoPosts, err := item.GetMemoPosts(postTxHashes)
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
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for set name resolver: %s", obj.TxHash)
	}
	return tx, nil
}

func (r *setNameResolver) Lock(ctx context.Context, obj *model.SetName) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for set name resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *setPicResolver) Tx(ctx context.Context, obj *model.SetPic) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for set pic resolver: %s", obj.TxHash)
	}
	return tx, nil
}

func (r *setPicResolver) Lock(ctx context.Context, obj *model.SetPic) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for set pic resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *setProfileResolver) Tx(ctx context.Context, obj *model.SetProfile) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for set profile resolver: %s", obj.TxHash)
	}
	return tx, nil
}

func (r *setProfileResolver) Lock(ctx context.Context, obj *model.SetProfile) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for set profile resolver: %s", obj.TxHash)
	}
	return lock, nil
}

// Follow returns generated.FollowResolver implementation.
func (r *Resolver) Follow() generated.FollowResolver { return &followResolver{r} }

// Like returns generated.LikeResolver implementation.
func (r *Resolver) Like() generated.LikeResolver { return &likeResolver{r} }

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
type likeResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type profileResolver struct{ *Resolver }
type setNameResolver struct{ *Resolver }
type setPicResolver struct{ *Resolver }
type setProfileResolver struct{ *Resolver }
