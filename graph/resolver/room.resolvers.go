package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/graph/load"
	"github.com/memocash/index/graph/model"
)

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
	tx, err := load.GetTx(ctx, obj.TxHash)
	if err != nil {
		return nil, jerr.Getf(err, "error getting tx from loader for room follow resolver: %s", obj.TxHash)
	}
	return tx, nil
}

// RoomFollow returns generated.RoomFollowResolver implementation.
func (r *Resolver) RoomFollow() generated.RoomFollowResolver { return &roomFollowResolver{r} }

type roomFollowResolver struct{ *Resolver }
