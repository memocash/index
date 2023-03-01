package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
)

// Output is the resolver for the output field.
func (r *slpBatonResolver) Output(ctx context.Context, obj *model.SlpBaton) (*model.TxOutput, error) {
	panic(fmt.Errorf("not implemented: Output - output"))
}

// Genesis is the resolver for the genesis field.
func (r *slpBatonResolver) Genesis(ctx context.Context, obj *model.SlpBaton) (*model.SlpGenesis, error) {
	panic(fmt.Errorf("not implemented: Genesis - genesis"))
}

// Tx is the resolver for the tx field.
func (r *slpGenesisResolver) Tx(ctx context.Context, obj *model.SlpGenesis) (*model.Tx, error) {
	panic(fmt.Errorf("not implemented: Tx - tx"))
}

// Output is the resolver for the output field.
func (r *slpGenesisResolver) Output(ctx context.Context, obj *model.SlpGenesis) (*model.SlpOutput, error) {
	panic(fmt.Errorf("not implemented: Output - output"))
}

// Baton is the resolver for the baton field.
func (r *slpGenesisResolver) Baton(ctx context.Context, obj *model.SlpGenesis) (*model.SlpBaton, error) {
	panic(fmt.Errorf("not implemented: Baton - baton"))
}

// Output is the resolver for the output field.
func (r *slpOutputResolver) Output(ctx context.Context, obj *model.SlpOutput) (*model.TxOutput, error) {
	panic(fmt.Errorf("not implemented: Output - output"))
}

// Genesis is the resolver for the genesis field.
func (r *slpOutputResolver) Genesis(ctx context.Context, obj *model.SlpOutput) (*model.SlpGenesis, error) {
	panic(fmt.Errorf("not implemented: Genesis - genesis"))
}

// SlpBaton returns generated.SlpBatonResolver implementation.
func (r *Resolver) SlpBaton() generated.SlpBatonResolver { return &slpBatonResolver{r} }

// SlpGenesis returns generated.SlpGenesisResolver implementation.
func (r *Resolver) SlpGenesis() generated.SlpGenesisResolver { return &slpGenesisResolver{r} }

// SlpOutput returns generated.SlpOutputResolver implementation.
func (r *Resolver) SlpOutput() generated.SlpOutputResolver { return &slpOutputResolver{r} }

type slpBatonResolver struct{ *Resolver }
type slpGenesisResolver struct{ *Resolver }
type slpOutputResolver struct{ *Resolver }
