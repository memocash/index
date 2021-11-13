package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/admin/graph/generated"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/node/obj/get"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	chainHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, jerr.Get("error getting tx hash from hash", err)
	}
	txHash := chainHash.CloneBytes()
	txHashString := chainHash.String()
	preloads := GetPreloads(ctx)
	var raw []byte
	if jutil.StringsInSlice([]string{"raw", "inputs", "outputs"}, preloads) {
		txBlocks, err := item.GetSingleTxBlocks(txHash)
		if err != nil {
			return nil, jerr.Get("error getting tx blocks from items", err)
		}
		if len(txBlocks) == 0 {
			mempoolTxRaw, err := item.GetMempoolTxRawByHash(txHash)
			if err != nil {
				return nil, jerr.Get("error getting mempool tx raw", err)
			}
			raw = mempoolTxRaw.Raw
		} else {
			txRaw, err := item.GetRawBlockTxByHash(txBlocks[0].BlockHash, txHash)
			if err != nil {
				return nil, jerr.Get("error getting block tx by hash", err)
			}
			raw = txRaw.Raw
		}
	}
	return &model.Tx{
		Hash: txHashString,
		Raw:  hex.EncodeToString(raw),
	}, nil
}

func (r *queryResolver) Address(ctx context.Context, address string) (*model.Lock, error) {
	balance, err := get.NewBalanceFromAddress(address)
	if err != nil {
		return nil, jerr.Get("error getting address from string for balance", err)
	}
	if err := balance.GetBalanceByUtxos(); err != nil {
		return nil, jerr.Get("error getting address balance from network", err)
	}
	return &model.Lock{
		Hash:    hex.EncodeToString(script.GetLockHash(balance.LockScript)),
		Address: balance.Address,
		Balance: balance.Balance,
	}, nil
}

func (r *queryResolver) Block(ctx context.Context, hash string) (*model.Block, error) {
	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, jerr.Get("error parsing block hash for block query resolver", err)
	}
	blockHeight, err := item.GetBlockHeight(blockHash.CloneBytes())
	if err != nil {
		return nil, jerr.Get("error getting block height for query resolver", err)
	}
	height := int(blockHeight.Height)
	return &model.Block{
		Hash:      hs.GetTxString(blockHeight.BlockHash),
		Timestamp: model.Date{},
		Height:    &height,
	}, nil
}

func (r *queryResolver) Blocks(ctx context.Context, newest *bool) ([]*model.Block, error) {
	heightBlocks, err := item.GetHeightBlocksAll(0, false)
	if err != nil {
		return nil, jerr.Get("error getting height blocks for query", err)
	}
	var modelBlocks = make([]*model.Block, len(heightBlocks))
	for i := range heightBlocks {
		var height = int(heightBlocks[i].Height)
		modelBlocks[i] = &model.Block{
			Hash:   hs.GetTxString(heightBlocks[i].BlockHash),
			Height: &height,
		}
	}
	return modelBlocks, nil
}

func (r *queryResolver) DoubleSpends(ctx context.Context) ([]*model.DoubleSpend, error) {
	doubleSpends, err := item.GetDoubleSpendOutputs(nil, client.DefaultLimit)
	if err != nil {
		return nil, jerr.Get("error getting double spend outputs", err)
	}
	var modelDoubleSpends = make([]*model.DoubleSpend, len(doubleSpends))
	for i := range doubleSpends {
		modelDoubleSpends[i] = &model.DoubleSpend{
			Hash:  hs.GetTxString(doubleSpends[i].TxHash),
			Index: doubleSpends[i].Index,
		}
	}
	return modelDoubleSpends, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
