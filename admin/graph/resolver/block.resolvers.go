package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
)

// Txs is the resolver for the txs field.
func (r *blockResolver) Txs(ctx context.Context, obj *model.Block, start *string) ([]*model.Tx, error) {
	blockHash, err := chainhash.NewHashFromStr(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error parsing block hash for block txs resolver", err)
	}
	var startUid []byte
	if start != nil {
		startTx, err := chainhash.NewHashFromStr(*start)
		if err != nil {
			return nil, jerr.Get("error decoding start tx hash for block", err)
		}
		startUid = chain.GetBlockTxUid(blockHash[:], startTx[:])
	}
	blockTxs, err := chain.GetBlockTxes(chain.BlockTxesRequest{
		BlockHash: blockHash[:],
		StartUid:  startUid,
		Limit:     client.DefaultLimit,
	})
	if err != nil {
		return nil, jerr.Get("error getting block transactions for hash", err)
	}
	var modelTxs = make([]*model.Tx, len(blockTxs))
	for i := range blockTxs {
		modelTxs[i] = &model.Tx{
			Index: blockTxs[i].Index,
			Hash:  chainhash.Hash(blockTxs[i].TxHash).String(),
		}
	}
	return modelTxs, nil
}

// Block returns generated.BlockResolver implementation.
func (r *Resolver) Block() generated.BlockResolver { return &blockResolver{r} }

type blockResolver struct{ *Resolver }
