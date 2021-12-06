package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

func (r *blockResolver) Txs(ctx context.Context, obj *model.Block) ([]*model.Tx, error) {
	blockHash, err := chainhash.NewHashFromStr(obj.Hash)
	if err != nil {
		return nil, jerr.Get("error parsing block hash for block txs resolver", err)
	}
	blockTxs, err := item.GetBlockTxes(item.BlockTxesRequest{
		BlockHash: blockHash.CloneBytes(),
	})
	if err != nil {
		return nil, jerr.Get("error getting block transactions for hash", err)
	}
	var modelTxs = make([]*model.Tx, len(blockTxs))
	for i := range blockTxs {
		modelTxs[i] = &model.Tx{
			Hash: hs.GetTxString(blockTxs[i].TxHash),
		}
	}
	return modelTxs, nil
}

// Block returns generated.BlockResolver implementation.
func (r *Resolver) Block() generated.BlockResolver { return &blockResolver{r} }

type blockResolver struct{ *Resolver }
