package load

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func GetBlock(ctx context.Context) *dataloader.BlockLoader {
	if !HasFieldAny(ctx, []string{"size", "tx_count"}) {
		return blockNoInfo
	}
	return blockWithInfo
}

var blockNoInfo = dataloader.NewBlockLoader(dataloader.BlockLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []string) ([][]*model.Block, []error) {
		modelBlocks, errs := block(keys, false)
		if errs != nil {
			return nil, errs
		}
		return modelBlocks, nil
	},
})

var blockWithInfo = dataloader.NewBlockLoader(dataloader.BlockLoaderConfig{
	Wait: defaultWait,
	Fetch: func(keys []string) ([][]*model.Block, []error) {
		modelBlocks, errs := block(keys, true)
		if errs != nil {
			return nil, errs
		}
		return modelBlocks, nil
	},
})

func block(keys []string, withInfo bool) ([][]*model.Block, []error) {
	var txHashes = make([][32]byte, len(keys))
	for i := range keys {
		hash, err := chainhash.NewHashFromStr(keys[i])
		if err != nil {
			return nil, []error{jerr.Get("error getting tx hash from string for block loader", err)}
		}
		txHashes[i] = *hash
	}
	txBlocks, err := chain.GetTxBlocks(txHashes)
	if err != nil {
		return nil, []error{jerr.Get("error getting blocks for tx for block loader", err)}
	}
	var blockHashes = make([][32]byte, len(txBlocks))
	for i := range txBlocks {
		blockHashes[i] = txBlocks[i].BlockHash
	}
	blocks, err := chain.GetBlocks(blockHashes)
	if err != nil {
		return nil, []error{jerr.Get("error getting blocks for block loader", err)}
	}
	blockHeights, err := chain.GetBlockHeights(blockHashes)
	if err != nil {
		return nil, []error{jerr.Get("error getting block heights for block loader", err)}
	}
	var blockInfos []*chain.BlockInfo
	if withInfo {
		if blockInfos, err = chain.GetBlockInfos(blockHashes); err != nil {
			return nil, []error{jerr.Get("error getting block infos for block loader", err)}
		}
	}
	var modelBlocks = make([][]*model.Block, len(txHashes))
	for i := range txHashes {
		for _, txBlock := range txBlocks {
			if txBlock.TxHash != txHashes[i] {
				continue
			}
			var modelBlock = &model.Block{
				Hash: txBlock.BlockHash,
			}
			for _, block := range blocks {
				if block.Hash != txBlock.BlockHash {
					continue
				}
				blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
				if err != nil {
					return nil, []error{jerr.Get("error getting block header from raw for block loader", err)}
				}
				modelBlock.Timestamp = model.Date(blockHeader.Timestamp)
				for _, blockHeight := range blockHeights {
					if blockHeight.BlockHash == block.Hash {
						height := int(blockHeight.Height)
						modelBlock.Height = &height
					}
				}
			}
			for _, blockInfo := range blockInfos {
				if blockInfo.BlockHash != txBlock.BlockHash {
					continue
				}
				modelBlock.Size = blockInfo.Size
				modelBlock.TxCount = blockInfo.TxCount
			}
			modelBlocks[i] = append(modelBlocks[i], modelBlock)
		}
	}
	return modelBlocks, nil
}
