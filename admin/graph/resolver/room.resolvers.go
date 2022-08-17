package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

func (r *roomResolver) Posts(ctx context.Context, obj *model.Room, start *int) ([]*model.Post, error) {
	roomHeightPosts, err := memo.GetRoomHeightPosts(ctx, obj.Name)
	if err != nil {
		return nil, jerr.Get("error getting room height posts for room resolver", err)
	}
	var txHashes = make([][]byte, len(roomHeightPosts))
	for i := range roomHeightPosts {
		txHashes[i] = roomHeightPosts[i].TxHash
	}
	memoPosts, err := memo.GetPosts(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting posts for room resolver", err)
	}
	var posts = make([]*model.Post, len(memoPosts))
	for i := range memoPosts {
		posts[i] = &model.Post{
			TxHash:   hs.GetTxString(memoPosts[i].TxHash),
			LockHash: hex.EncodeToString(memoPosts[i].LockHash),
			Text:     memoPosts[i].Post,
		}
	}
	return posts, nil
}

func (r *roomResolver) Followers(ctx context.Context, obj *model.Room, start *int) ([]*model.RoomFollow, error) {
	lockRoomFollows, err := memo.GetRoomHeightFollows(ctx, obj.Name)
	if err != nil {
		return nil, jerr.Get("error getting room height follows for followers in room resolver", err)
	}
	var roomFollows = make([]*model.RoomFollow, len(lockRoomFollows))
	for i := range lockRoomFollows {
		roomFollows[i] = &model.RoomFollow{
			Name:     obj.Name,
			LockHash: hex.EncodeToString(lockRoomFollows[i].LockHash),
			Unfollow: lockRoomFollows[i].Unfollow,
			TxHash:   hs.GetTxString(lockRoomFollows[i].TxHash),
		}
	}
	return roomFollows, nil
}

func (r *roomFollowResolver) Room(ctx context.Context, obj *model.RoomFollow) (*model.Room, error) {
	return &model.Room{Name: obj.Name}, nil
}

func (r *roomFollowResolver) Lock(ctx context.Context, obj *model.RoomFollow) (*model.Lock, error) {
	lock, err := LockLoader(ctx, obj.LockHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for room follow resolver: %s", obj.TxHash)
	}
	return lock, nil
}

func (r *roomFollowResolver) Tx(ctx context.Context, obj *model.RoomFollow) (*model.Tx, error) {
	tx, err := TxLoader(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for room follow resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// Room returns generated.RoomResolver implementation.
func (r *Resolver) Room() generated.RoomResolver { return &roomResolver{r} }

// RoomFollow returns generated.RoomFollowResolver implementation.
func (r *Resolver) RoomFollow() generated.RoomFollowResolver { return &roomFollowResolver{r} }

type roomResolver struct{ *Resolver }
type roomFollowResolver struct{ *Resolver }
