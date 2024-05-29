package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/graph/load"
	"github.com/memocash/index/graph/model"
)

// Tx is the resolver for the tx field.
func (r *setNameResolver) Tx(ctx context.Context, obj *model.SetName) (*model.Tx, error) {
	tx, err := load.GetTx(ctx, obj.TxHash)
	if err != nil {
		return nil, fmt.Errorf("error getting tx from loader for set name resolver: %s; %w", obj.TxHash, err)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *setNameResolver) Lock(ctx context.Context, obj *model.SetName) (*model.Lock, error) {
	lock, err := load.GetLock(ctx, obj.Address)
	if err != nil {
		return nil, fmt.Errorf("error getting lock from loader for set name resolver: %s %x; %w", obj.TxHash, obj.Address, err)
	}
	return lock, nil
}

// Tx is the resolver for the tx field.
func (r *setPicResolver) Tx(ctx context.Context, obj *model.SetPic) (*model.Tx, error) {
	tx, err := load.GetTx(ctx, obj.TxHash)
	if err != nil {
		return nil, fmt.Errorf("error getting tx from loader for set pic resolver: %s; %w", obj.TxHash, err)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *setPicResolver) Lock(ctx context.Context, obj *model.SetPic) (*model.Lock, error) {
	lock, err := load.GetLock(ctx, obj.Address)
	if err != nil {
		return nil, fmt.Errorf("error getting lock from loader for set pic resolver: %s %x; %w", obj.TxHash, obj.Address, err)
	}
	return lock, nil
}

// Tx is the resolver for the tx field.
func (r *setProfileResolver) Tx(ctx context.Context, obj *model.SetProfile) (*model.Tx, error) {
	tx, err := load.GetTx(ctx, obj.TxHash)
	if err != nil {
		return nil, fmt.Errorf("error getting tx from loader for set profile resolver: %s; %w", obj.TxHash, err)
	}
	return tx, nil
}

// Lock is the resolver for the lock field.
func (r *setProfileResolver) Lock(ctx context.Context, obj *model.SetProfile) (*model.Lock, error) {
	lock, err := load.GetLock(ctx, obj.Address)
	if err != nil {
		return nil, fmt.Errorf("error getting lock from loader for set profile resolver: %s; %w", obj.TxHash, err)
	}
	return lock, nil
}

// SetName returns generated.SetNameResolver implementation.
func (r *Resolver) SetName() generated.SetNameResolver { return &setNameResolver{r} }

// SetPic returns generated.SetPicResolver implementation.
func (r *Resolver) SetPic() generated.SetPicResolver { return &setPicResolver{r} }

// SetProfile returns generated.SetProfileResolver implementation.
func (r *Resolver) SetProfile() generated.SetProfileResolver { return &setProfileResolver{r} }

type setNameResolver struct{ *Resolver }
type setPicResolver struct{ *Resolver }
type setProfileResolver struct{ *Resolver }
