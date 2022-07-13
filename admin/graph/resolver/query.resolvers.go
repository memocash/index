package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	txHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, jerr.Get("error getting tx hash from hash", err)
	}
	txHashString := txHash.String()
	preloads := GetPreloads(ctx)
	var raw string
	if jutil.StringsInSlice([]string{"raw", "inputs", "outputs"}, preloads) {
		if raw, err = dataloader.NewTxRawLoader(txRawLoaderConfig).Load(txHashString); err != nil {
			return nil, jerr.Get("error getting tx raw from dataloader for tx query resolver", err)
		}
	}
	return &model.Tx{
		Hash: txHashString,
		Raw:  raw,
	}, nil
}

func (r *queryResolver) Txs(ctx context.Context, hashes []string) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
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

func (r *queryResolver) Addresses(ctx context.Context, addresses []string) ([]*model.Lock, error) {
	var locks []*model.Lock
	for _, address := range addresses {
		balance, err := get.NewBalanceFromAddress(address)
		if err != nil {
			return nil, jerr.Get("error getting address from string for balances", err)
		}
		if err := balance.GetBalance(); err != nil {
			return nil, jerr.Get("error getting address balance from network (multi)", err)
		}
		locks = append(locks, &model.Lock{
			Hash:    hex.EncodeToString(script.GetLockHash(balance.LockScript)),
			Address: balance.Address,
			Balance: balance.Balance,
		})
	}
	return locks, nil
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
	block, err := item.GetBlock(blockHash.CloneBytes())
	if err != nil {
		return nil, jerr.Get("error getting raw block", err)
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return nil, jerr.Get("error getting block header from raw", err)
	}
	height := int(blockHeight.Height)
	return &model.Block{
		Hash:      hs.GetTxString(blockHeight.BlockHash),
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
		Raw:       hex.EncodeToString(block.Raw),
	}, nil
}

func (r *queryResolver) BlockNewest(ctx context.Context) (*model.Block, error) {
	heightBlock, err := item.GetRecentHeightBlock()
	if err != nil {
		return nil, jerr.Get("error getting recent height block for query", err)
	}
	if heightBlock == nil {
		return nil, nil
	}
	block, err := item.GetBlock(heightBlock.BlockHash)
	if err != nil {
		return nil, jerr.Get("error getting raw block", err)
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return nil, jerr.Get("error getting block header from raw", err)
	}
	height := int(heightBlock.Height)
	return &model.Block{
		Hash:      hs.GetTxString(heightBlock.BlockHash),
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
	}, nil
}

func (r *queryResolver) Blocks(ctx context.Context, newest *bool, start *uint32) ([]*model.Block, error) {
	var startInt int64
	if start != nil {
		startInt = int64(*start)
	}
	var newestBool bool
	if newest != nil {
		newestBool = *newest
	}
	heightBlocks, err := item.GetHeightBlocksAllDefault(startInt, false, newestBool)
	if err != nil {
		return nil, jerr.Get("error getting height blocks for query", err)
	}
	var blockHashes = make([][]byte, len(heightBlocks))
	for i := range heightBlocks {
		blockHashes[i] = heightBlocks[i].BlockHash
	}
	blocks, err := item.GetBlocks(blockHashes)
	if err != nil {
		return nil, jerr.Get("error getting raw blocks", err)
	}
	var modelBlocks = make([]*model.Block, len(heightBlocks))
	for i := range heightBlocks {
		var height = int(heightBlocks[i].Height)
		modelBlocks[i] = &model.Block{
			Hash:   hs.GetTxString(heightBlocks[i].BlockHash),
			Height: &height,
		}
		for _, block := range blocks {
			if bytes.Equal(block.Hash, heightBlocks[i].BlockHash) {
				blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
				if err != nil {
					return nil, jerr.Get("error getting block header from raw", err)
				}
				modelBlocks[i].Timestamp = model.Date(blockHeader.Timestamp)
			}
		}
	}
	return modelBlocks, nil
}

func (r *queryResolver) DoubleSpends(ctx context.Context, newest *bool, start *model.Date) ([]*model.DoubleSpend, error) {
	var startTime time.Time
	if start != nil {
		startTime = time.Time(*start)
	}
	var newestBool bool
	if newest != nil {
		newestBool = *newest
	}
	doubleSpendSeens, err := item.GetDoubleSpendSeensAllLimit(startTime, client.DefaultLimit, newestBool)
	if err != nil {
		return nil, jerr.Get("error getting double spend outputs", err)
	}
	var modelDoubleSpends = make([]*model.DoubleSpend, len(doubleSpendSeens))
	for i := range doubleSpendSeens {
		modelDoubleSpends[i] = &model.DoubleSpend{
			Hash:      hs.GetTxString(doubleSpendSeens[i].TxHash),
			Index:     doubleSpendSeens[i].Index,
			Timestamp: model.Date(doubleSpendSeens[i].Timestamp),
		}
	}
	return modelDoubleSpends, nil
}

func (r *subscriptionResolver) Address(ctx context.Context, address string) (<-chan *model.Tx, error) {
	lockScript, err := get.LockScriptFromAddress(wallet.GetAddressFromString(address))
	if err != nil {
		return nil, jerr.Get("error getting lock script for address subscription", err)
	}
	ctx, cancel := context.WithCancel(ctx)
	lockHeightOutputsListener, err := item.ListenMempoolLockHeightOutputs(ctx, script.GetLockHash(lockScript))
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting lock height outputs listener for address subscription", err)
	}
	lockHeightInputsListener, err := item.ListenMempoolLockHeightOutputInputs(ctx, script.GetLockHash(lockScript))
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting lock height inputs listener for address subscription", err)
	}
	var txChan = make(chan *model.Tx)
	go func() {
		defer func() {
			txChan <- nil
			close(lockHeightOutputsListener)
			close(lockHeightInputsListener)
			cancel()
		}()
		for {
			select {
			case lockHeightOutput := <-lockHeightOutputsListener:
				if lockHeightOutput == nil {
					jlog.Log("nil lock height output, closing address subscription")
					return
				}
				txRaw, err := item.GetMempoolTxRawByHash(lockHeightOutput.Hash)
				if err != nil {
					jerr.Get("error getting mempool tx raw for address subscription output", err).Print()
					return
				}
				txChan <- &model.Tx{
					Hash: hs.GetTxString(lockHeightOutput.Hash),
					Raw:  hex.EncodeToString(txRaw.Raw),
				}
			case lockHeightOutputInput := <-lockHeightInputsListener:
				if lockHeightOutputInput == nil {
					jlog.Log("nil lock height output input, closing address subscription")
					return
				}
				txRaw, err := item.GetMempoolTxRawByHash(lockHeightOutputInput.Hash)
				if err != nil {
					jerr.Get("error getting mempool tx raw for address subscription input", err).Print()
					return
				}
				txChan <- &model.Tx{
					Hash: hs.GetTxString(lockHeightOutputInput.Hash),
					Raw:  hex.EncodeToString(txRaw.Raw),
				}
			}

		}
	}()
	return txChan, nil
}

func (r *subscriptionResolver) Blocks(ctx context.Context) (<-chan *model.Block, error) {
	blockHeightListener, err := item.ListenBlockHeights(ctx)
	if err != nil {
		return nil, jerr.Get("error getting block height listener for subscription", err)
	}
	var blockChan = make(chan *model.Block)
	go func() {
		defer func() { blockChan <- nil }()
		for {
			var blockHeight *item.BlockHeight
			select {
			case <-ctx.Done():
				jerr.Get("error blocks subscription context done", ctx.Err()).Print()
				return
			case blockHeight = <-blockHeightListener:
			}
			if blockHeight == nil {
				jlog.Log("nil block height, closing subscription")
				return
			}
			block, err := item.GetBlock(blockHeight.BlockHash)
			if err != nil {
				jerr.Get("error getting block for block height subscription", err).Print()
				return
			}
			blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
			if err != nil {
				jerr.Get("error getting block header from raw", err).Print()
				return
			}
			height := int(blockHeight.Height)
			blockChan <- &model.Block{
				Hash:      hs.GetTxString(blockHeight.BlockHash),
				Timestamp: model.Date(blockHeader.Timestamp),
				Height:    &height,
			}
		}
	}()
	return blockChan, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
