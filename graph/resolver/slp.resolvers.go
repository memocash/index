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

// Output is the resolver for the output field.
func (r *slpOutputResolver) Output(ctx context.Context, obj *model.SlpOutput) (*model.TxOutput, error) {
	txOutput, err := load.GetTxOutput(ctx, obj.Hash, obj.Index)
	if err != nil {
		return nil, jerr.Get("error getting tx output for slp output from loader", err)
	}
	return txOutput, nil
}

// SlpOutput returns generated.SlpOutputResolver implementation.
func (r *Resolver) SlpOutput() generated.SlpOutputResolver { return &slpOutputResolver{r} }

type slpOutputResolver struct{ *Resolver }
