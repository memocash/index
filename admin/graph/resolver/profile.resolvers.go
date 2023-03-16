package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Tx is the resolver for the tx field.
func (r *followResolver) Tx(ctx context.Context, obj *model.Follow) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for follow resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *followResolver) Lock(ctx context.Context, obj *model.Follow) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for follow resolver: %s", obj.TxHash)
	}
	return lock, nil
}

// FollowLock is the resolver for the follow_lock field.
func (r *followResolver) FollowLock(ctx context.Context, obj *model.Follow) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.FollowAddress)
	if err != nil {
		return nil, jerr.Getf(err, "error getting follow lock from loader for follow resolver: %s", obj.TxHash)
	}
	return lock, nil
}

// Tx is the resolver for the tx field.
func (r *likeResolver) Tx(ctx context.Context, obj *model.Like) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for like resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *likeResolver) Lock(ctx context.Context, obj *model.Like) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for like resolver: %s %x", obj.TxHash, obj.Address)
	}
	return lock, nil
}

// Post is the resolver for the post field.
func (r *likeResolver) Post(ctx context.Context, obj *model.Like) (*model.Post, error) {
	post, err := load.Post.Load(obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting post from post dataloader for like resolver: %s", obj.Address)
	}
	return post, nil
}

// Tx is the resolver for the tx field.
func (r *postResolver) Tx(ctx context.Context, obj *model.Post) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for post resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *postResolver) Lock(ctx context.Context, obj *model.Post) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for post resolver: %s %x", obj.TxHash, obj.Address)
	}
	return lock, nil
}

// Likes is the resolver for the likes field.
func (r *postResolver) Likes(ctx context.Context, obj *model.Post) ([]*model.Like, error) {
	postTxHash, err := chainhash.NewHashFromStr(obj.TxHash)
	if err != nil {
		return nil, jerr.Get("error parsing tx hash for likes for post resolver", err)
	}
	memoPostLikes, err := memo.GetPostLikes([][32]byte{*postTxHash})
	if err != nil {
		return nil, jerr.Get("error getting memo post likeds for post resolver", err)
	}
	var likeTxHashes = make([][32]byte, len(memoPostLikes))
	for i := range memoPostLikes {
		likeTxHashes[i] = memoPostLikes[i].LikeTxHash
	}
	memoLikeTips, err := memo.GetLikeTips(likeTxHashes)
	if err != nil {
		return nil, jerr.Get("error getting memo like tips for post resolver", err)
	}
	var likes = make([]*model.Like, len(memoPostLikes))
	for i := range memoPostLikes {
		var tip int64
		for j := range memoLikeTips {
			if memoLikeTips[j].LikeTxHash == memoPostLikes[i].LikeTxHash {
				tip = memoLikeTips[j].Tip
				memoLikeTips = append(memoLikeTips[:j], memoLikeTips[j+1:]...)
				break
			}
		}
		likes[i] = &model.Like{
			TxHash:     chainhash.Hash(memoPostLikes[i].LikeTxHash).String(),
			PostTxHash: chainhash.Hash(memoPostLikes[i].PostTxHash).String(),
			Address:    wallet.Addr(memoPostLikes[i].Addr).String(),
			Tip:        tip,
		}
	}
	return likes, nil
}

// Parent is the resolver for the parent field.
func (r *postResolver) Parent(ctx context.Context, obj *model.Post) (*model.Post, error) {
	postTxHash, err := chainhash.NewHashFromStr(obj.TxHash)
	if err != nil {
		return nil, jerr.Get("error parsing tx hash for likes for post resolver", err)
	}
	postParent, err := memo.GetPostParent(ctx, *postTxHash)
	if err != nil {
		return nil, jerr.Get("error getting memo post parent for post resolver", err)
	}
	if postParent == nil {
		return nil, nil
	}
	post, err := load.Post.Load(chainhash.Hash(postParent.ParentTxHash).String())
	if err != nil {
		if load.IsPostNotFoundError(err) {
			return nil, nil
		}
		return nil, jerr.Getf(err, "error getting from post dataloader for post parent resolver: %s", obj.Address)
	}
	return post, nil
}

// Replies is the resolver for the replies field.
func (r *postResolver) Replies(ctx context.Context, obj *model.Post) ([]*model.Post, error) {
	postTxHash, err := chainhash.NewHashFromStr(obj.TxHash)
	if err != nil {
		return nil, jerr.Get("error parsing tx hash for likes for post resolver", err)
	}
	postChildren, err := memo.GetPostChildren(ctx, *postTxHash)
	if err != nil {
		return nil, jerr.Get("error getting memo post children for post resolver", err)
	}
	var childrenTxHashes = make([]string, len(postChildren))
	for i := range postChildren {
		childrenTxHashes[i] = chainhash.Hash(postChildren[i].ChildTxHash).String()
	}
	replies, errs := load.Post.LoadAll(childrenTxHashes)
	for _, err := range errs {
		if err != nil {
			return nil, jerr.Getf(err, "error getting from post dataloader for post reply resolver: %s", obj.TxHash)
		}
	}
	return replies, nil
}

// Room is the resolver for the room field.
func (r *postResolver) Room(ctx context.Context, obj *model.Post) (*model.Room, error) {
	postTxHash, err := chainhash.NewHashFromStr(obj.TxHash)
	if err != nil {
		return nil, jerr.Get("error parsing tx hash for room for post resolver", err)
	}

	postRoom, err := memo.GetPostRoom(ctx, postTxHash.CloneBytes())
	if err != nil {
		return nil, jerr.Get("error getting memo post room for post resolver", err)
	}
	return &model.Room{
		Name: postRoom.Room,
	}, nil
}

// Lock is the resolver for the lock field.
func (r *profileResolver) Lock(ctx context.Context, obj *model.Profile) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting addr from loader for profile resolver: %s", obj.Address)
	}
	return lock, nil
}

// Following is the resolver for the following field.
func (r *profileResolver) Following(ctx context.Context, obj *model.Profile, start *model.Date) ([]*model.Follow, error) {
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting address from string for profile following resolver: %s", obj.Address)
	}
	var startTime time.Time
	if start != nil {
		startTime = time.Time(*start)
	}
	addrMemoFollows, err := memo.GetAddrFollowsSingle(ctx, *address, startTime)
	if err != nil {
		return nil, jerr.Get("error getting address memo follows for address", err)
	}
	var follows []*model.Follow
	for _, addrMemoFollow := range addrMemoFollows {
		follows = append(follows, &model.Follow{
			TxHash:        chainhash.Hash(addrMemoFollow.TxHash).String(),
			Address:       wallet.Addr(addrMemoFollow.Addr).String(),
			FollowAddress: wallet.Addr(addrMemoFollow.FollowAddr).String(),
			Unfollow:      addrMemoFollow.Unfollow,
		})
	}
	return follows, nil
}

// Followers is the resolver for the followers field.
func (r *profileResolver) Followers(ctx context.Context, obj *model.Profile, start *model.Date) ([]*model.Follow, error) {
	address, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting address from string for profile followers resolver: %s", obj.Address)
	}
	var startTime time.Time
	if start != nil {
		startTime = time.Time(*start)
	}
	addrMemoFolloweds, err := memo.GetAddrFollowedsSingle(ctx, *address, startTime)
	if err != nil {
		return nil, jerr.Getf(err, "error getting addr memo follows for address: %s", obj.Address)
	}
	var follows []*model.Follow
	for _, addrMemoFollowed := range addrMemoFolloweds {
		follows = append(follows, &model.Follow{
			TxHash:        chainhash.Hash(addrMemoFollowed.TxHash).String(),
			Address:       wallet.Addr(addrMemoFollowed.Addr).String(),
			FollowAddress: wallet.Addr(addrMemoFollowed.FollowAddr).String(),
			Unfollow:      addrMemoFollowed.Unfollow,
		})
	}
	return follows, nil
}

// Posts is the resolver for the posts field.
func (r *profileResolver) Posts(ctx context.Context, obj *model.Profile, start *model.Date, newest *bool) ([]*model.Post, error) {
	addr, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error decoding address for profile resolver: %s", obj.Address)
	}
	addrMemoPosts, err := memo.GetAddrPosts(ctx, [][25]byte{*addr}, newest != nil && *newest)
	if err != nil {
		return nil, jerr.Getf(err, "error getting addr memo posts for profile resolver: %s", obj.Address)
	}
	var postTxHashes = make([][32]byte, len(addrMemoPosts))
	for i := range addrMemoPosts {
		postTxHashes[i] = addrMemoPosts[i].TxHash
	}
	memoPosts, err := memo.GetPosts(postTxHashes)
	if err != nil {
		return nil, jerr.Get("error getting memo posts for profile resolver", err)
	}
	var posts = make([]*model.Post, len(memoPosts))
	for i, memoPost := range memoPosts {
		posts[i] = &model.Post{
			TxHash:  chainhash.Hash(memoPost.TxHash).String(),
			Address: wallet.Addr(memoPost.Addr).String(),
			Text:    memoPost.Post,
		}
	}
	return posts, nil
}

// Rooms is the resolver for the rooms field.
func (r *profileResolver) Rooms(ctx context.Context, obj *model.Profile, start *model.Date) ([]*model.RoomFollow, error) {
	addr, err := wallet.GetAddrFromString(obj.Address)
	if err != nil {
		return nil, jerr.Get("error decoding lock for room follows in profile resolver", err)
	}
	lockRoomFollows, err := memo.GetAddrRoomFollows(ctx, [][25]byte{*addr})
	var roomFollows = make([]*model.RoomFollow, len(lockRoomFollows))
	for i := range lockRoomFollows {
		roomFollows[i] = &model.RoomFollow{
			Name:     lockRoomFollows[i].Room,
			Address:  wallet.Addr(lockRoomFollows[i].Addr).String(),
			Unfollow: lockRoomFollows[i].Unfollow,
			TxHash:   chainhash.Hash(lockRoomFollows[i].TxHash).String(),
		}
	}
	return roomFollows, nil
}

// Tx is the resolver for the tx field.
func (r *setNameResolver) Tx(ctx context.Context, obj *model.SetName) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for set name resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *setNameResolver) Lock(ctx context.Context, obj *model.SetName) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for set name resolver: %s %x", obj.TxHash, obj.Address)
	}
	return lock, nil
}

// Tx is the resolver for the tx field.
func (r *setPicResolver) Tx(ctx context.Context, obj *model.SetPic) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for set pic resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *setPicResolver) Lock(ctx context.Context, obj *model.SetPic) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for set pic resolver: %s %x", obj.TxHash, obj.Address)
	}
	return lock, nil
}

// Tx is the resolver for the tx field.
func (r *setProfileResolver) Tx(ctx context.Context, obj *model.SetProfile) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for set profile resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *setProfileResolver) Lock(ctx context.Context, obj *model.SetProfile) (*model.Lock, error) {
	lock, err := load.Lock(ctx, obj.Address)
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
