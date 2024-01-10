package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
)

// Txs is the resolver for the txs field.
func (r *blockResolver) Txs(ctx context.Context, obj *model.Block, start *uint32) ([]*model.TxBlock, error) {
	var startIndex uint32
	if start != nil {
		startIndex = *start
	}
	blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
		BlockHash:  obj.Hash,
		StartIndex: startIndex,
		Limit:      client.DefaultLimit,
	})
	if err != nil {
		return nil, jerr.Get("error getting block transactions for hash", err)
	}
	var modelTxs = make([]*model.TxBlock, len(blockTxs))
	for i := range blockTxs {
		modelTxs[i] = &model.TxBlock{
			Index:  blockTxs[i].Index,
			TxHash: blockTxs[i].TxHash,
			Tx:     &model.Tx{Hash: blockTxs[i].TxHash},
		}
	}
	return modelTxs, nil
}

// Block returns generated.BlockResolver implementation.
func (r *Resolver) Block() generated.BlockResolver { return &blockResolver{r} }

type blockResolver struct{ *Resolver }
