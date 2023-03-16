package load

import (
	"context"
	"fmt"
	"github.com/graph-gophers/dataloader"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TxBlocksReader struct {
}

func (r *TxBlocksReader) GetTxBlocks(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	return getTxBlocks(ctx, keys, false)
}

func GetTxBlocks(ctx context.Context, txHash string) ([]*model.Block, error) {
	loaders := For(ctx)
	thunk := loaders.TxBlocksLoader.Load(ctx, dataloader.StringKey(txHash))
	result, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("error getting tx output inputs from loader; %w", err)
	}
	return result.([]*model.Block), nil
}

type TxBlocksWithInfoReader struct {
}

func (r *TxBlocksWithInfoReader) GetTxBlocks(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	return getTxBlocks(ctx, keys, true)
}

func GetTxBlocksWithInfo(ctx context.Context, txHash string) ([]*model.Block, error) {
	loaders := For(ctx)
	thunk := loaders.TxBlocksWithInfoLoader.Load(ctx, dataloader.StringKey(txHash))
	result, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("error getting tx output inputs with script from loader; %w", err)
	}
	return result.([]*model.Block), nil
}

func getTxBlocks(ctx context.Context, keys dataloader.Keys, withInfo bool) []*dataloader.Result {
	var results = make([]*dataloader.Result, len(keys))
	var txHashes [][32]byte
	for i := range keys {
		hash, err := chainhash.NewHashFromStr(keys[i].String())
		if err != nil {
			results[i] = &dataloader.Result{
				Error: fmt.Errorf("error getting tx hash for blocks dataloader: %s; %w", keys[i], err)}
			continue
		}
		txHashes = append(txHashes, *hash)
	}
	txBlocks, err := chain.GetTxBlocks(txHashes)
	if err != nil {
		return resultsError(results, fmt.Errorf("error getting tx blocks for tx for block loader; %w", err))
	}
	var blockHashes = make([][32]byte, len(txBlocks))
	for i := range txBlocks {
		blockHashes[i] = txBlocks[i].BlockHash
	}
	blocks, err := chain.GetBlocks(blockHashes)
	if err != nil {
		return resultsError(results, fmt.Errorf("error getting blocks for tx for block loader; %w", err))
	}
	blockHeights, err := chain.GetBlockHeights(blockHashes)
	if err != nil {
		return resultsError(results, fmt.Errorf("error getting block heights for tx for block loader; %w", err))
	}
	var blockInfos []*chain.BlockInfo
	if withInfo {
		if blockInfos, err = chain.GetBlockInfos(blockHashes); err != nil {
			return resultsError(results, fmt.Errorf("error getting block infos for tx for block loader; %w", err))
		}
	}
	var blocksByTxHash = make(map[string][]*model.Block)
	for _, txBlock := range txBlocks {
		var modelBlock = &model.Block{Hash: chainhash.Hash(txBlock.BlockHash).String()}
		for _, block := range blocks {
			if block.Hash == txBlock.BlockHash {
				blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
				if err != nil {
					return resultsError(results,
						fmt.Errorf("error getting block header from raw for block loader; %w", err))
				}
				modelBlock.Timestamp = model.Date(blockHeader.Timestamp)
				break
			}
		}
		for _, blockHeight := range blockHeights {
			if blockHeight.BlockHash == txBlock.BlockHash {
				height := int(blockHeight.Height)
				modelBlock.Height = &height
				break
			}
		}
		for _, blockInfo := range blockInfos {
			if blockInfo.BlockHash == txBlock.BlockHash {
				modelBlock.Size = blockInfo.Size
				modelBlock.TxCount = blockInfo.TxCount
				break
			}
		}
		var txHash = chainhash.Hash(txBlock.TxHash).String()
		blocksByTxHash[txHash] = append(blocksByTxHash[txHash], modelBlock)
	}
	for index, txHash := range keys {
		txHashBlocks, ok := blocksByTxHash[txHash.String()]
		if ok {
			results[index] = &dataloader.Result{Data: txHashBlocks}
		} else if results[index] == nil {
			results[index] = &dataloader.Result{Data: []*model.Block{}}
		}
	}
	return results
}
