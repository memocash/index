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

// Baton is the resolver for the baton field.
func (r *slpGenesisResolver) Baton(ctx context.Context, obj *model.SlpGenesis) (*model.SlpBaton, error) {
	var slpBaton = &model.SlpBaton{
		Hash:  obj.Hash,
		Index: obj.BatonIndex,
	}
	if err := load.AttachToSlpBatons(ctx, load.GetFields(ctx), []*model.SlpBaton{slpBaton}); err != nil {
		return nil, jerr.Get("error attaching slp baton for slp genesis from loader", err)
	}
	return slpBaton, nil
}

// Output is the resolver for the output field.
func (r *slpOutputResolver) Output(ctx context.Context, obj *model.SlpOutput) (*model.TxOutput, error) {
	txOutput, err := load.GetTxOutput(ctx, obj.Hash, obj.Index)
	if err != nil {
		return nil, jerr.Get("error getting tx output for slp output from loader", err)
	}
	return txOutput, nil
}

// SlpGenesis returns generated.SlpGenesisResolver implementation.
func (r *Resolver) SlpGenesis() generated.SlpGenesisResolver { return &slpGenesisResolver{r} }

// SlpOutput returns generated.SlpOutputResolver implementation.
func (r *Resolver) SlpOutput() generated.SlpOutputResolver { return &slpOutputResolver{r} }

type slpGenesisResolver struct{ *Resolver }
type slpOutputResolver struct{ *Resolver }
