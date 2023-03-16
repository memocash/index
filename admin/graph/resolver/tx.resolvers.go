package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
)

// Blocks is the resolver for the blocks field.
func (r *txResolver) Blocks(ctx context.Context, obj *model.Tx) ([]*model.Block, error) {
	blocks, err := load.GetBlock(ctx).Load(chainhash.Hash(obj.Hash).String())
	if err != nil {
		return nil, jerr.Get("error getting blocks for tx from loader", err)
	}
	return blocks, nil
}

// Tx returns generated.TxResolver implementation.
func (r *Resolver) Tx() generated.TxResolver { return &txResolver{r} }

type txResolver struct{ *Resolver }
