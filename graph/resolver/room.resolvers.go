package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/graph/load"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Followers is the resolver for the followers field.
func (r *roomResolver) Followers(ctx context.Context, obj *model.Room, start *int) ([]*model.RoomFollow, error) {
	lockRoomFollows, err := memo.GetRoomFollows(ctx, obj.Name)
	if err != nil {
		return nil, jerr.Get("error getting room height follows for followers in room resolver", err)
	}
	var roomFollows = make([]*model.RoomFollow, len(lockRoomFollows))
	for i := range lockRoomFollows {
		roomFollows[i] = &model.RoomFollow{
			Name:     obj.Name,
			Address:  wallet.Addr(lockRoomFollows[i].Addr).String(),
			Unfollow: lockRoomFollows[i].Unfollow,
			TxHash:   chainhash.Hash(lockRoomFollows[i].TxHash).String(),
		}
	}
	return roomFollows, nil
}

// Room is the resolver for the room field.
func (r *roomFollowResolver) Room(ctx context.Context, obj *model.RoomFollow) (*model.Room, error) {
	var room = &model.Room{Name: obj.Name}
	if err := load.AttachToMemoRooms(ctx, load.GetFields(ctx), []*model.Room{room}); err != nil {
		return nil, jerr.Getf(err, "error attaching to memo rooms for room follow resolver: %s", obj.TxHash)
	}
	return room, nil
}

// Lock is the resolver for the lock field.
func (r *roomFollowResolver) Lock(ctx context.Context, obj *model.RoomFollow) (*model.Lock, error) {
	lock, err := load.GetLock(ctx, obj.Address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for room follow resolver: %s", obj.TxHash)
	}
	return lock, nil
}

// Tx is the resolver for the tx field.
func (r *roomFollowResolver) Tx(ctx context.Context, obj *model.RoomFollow) (*model.Tx, error) {
	tx, err := load.GetTxByString(ctx, obj.TxHash)
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
