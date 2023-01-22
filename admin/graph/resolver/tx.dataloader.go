package resolver

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/act/tx_raw"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

var txLostLoaderConfig = dataloader.TxLostLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([]*model.TxLost, []error) {
		var txHashes = make([][]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			txHashes[i] = hash.CloneBytes()
		}
		txLosts, err := item.GetTxLosts(txHashes)
		if err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting tx losts", err)}
		}
		var modelTxLosts = make([]*model.TxLost, len(txHashes))
		for i := range txHashes {
			for _, txLost := range txLosts {
				if bytes.Equal(txLost.TxHash, txHashes[i]) {
					hash, err := chainhash.NewHash(txLost.TxHash)
					if err != nil {
						return nil, []error{jerr.Get("error parsing tx hash from tx lost", err)}
					}
					modelTxLosts[i] = &model.TxLost{
						Hash: hash.String(),
					}
					break
				}
			}
		}
		return modelTxLosts, nil
	},
}

var txSuspectLoaderConfig = dataloader.TxSuspectLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([]*model.TxSuspect, []error) {
		var txHashes = make([][]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			txHashes[i] = hash.CloneBytes()
		}
		txSuspects, err := item.GetTxSuspects(txHashes)
		if err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting tx suspects", err)}
		}
		var modelTxSuspects = make([]*model.TxSuspect, len(txHashes))
		for i := range txHashes {
			for _, txSuspect := range txSuspects {
				if bytes.Equal(txSuspect.TxHash, txHashes[i]) {
					hash, err := chainhash.NewHash(txSuspect.TxHash)
					if err != nil {
						return nil, []error{jerr.Get("error parsing tx hash from tx suspect", err)}
					}
					modelTxSuspects[i] = &model.TxSuspect{
						Hash: hash.String(),
					}
					break
				}
			}
		}
		return modelTxSuspects, nil
	},
}

var txSeenLoaderConfig = dataloader.TxSeenLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([]*model.Date, []error) {
		var txHashes = make([][32]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			txHashes[i] = *hash
		}
		txSeens, err := item.GetTxSeens(txHashes)
		if err != nil && !client.IsResourceUnavailableError(err) {
			return nil, []error{jerr.Get("error getting tx seens", err)}
		}
		var modelTxSeens = make([]*model.Date, len(txHashes))
		for i := range txHashes {
			for _, txSeen := range txSeens {
				if txSeen.TxHash == txHashes[i] {
					if modelTxSeens[i] == nil || time.Time(*modelTxSeens[i]).After(txSeen.Timestamp) {
						var modelDate = model.Date(txSeen.Timestamp)
						modelTxSeens[i] = &modelDate
					}
				}
			}
		}
		return modelTxSeens, nil
	},
}

var txRawLoaderConfig = dataloader.TxRawLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([]string, []error) {
		var txHashes = make([][32]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error parsing tx hash for raw loader", err)}
			}
			txHashes[i] = *hash
		}
		txRaws, err := tx_raw.Get(txHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting tx raws for tx dataloader", err)}
		}
		txRawStrings := make([]string, len(txRaws))
		for i := range txRaws {
			txRawStrings[i] = hex.EncodeToString(txRaws[i].Raw)
		}
		return txRawStrings, nil
	},
}

func GetBlockLoaderConfig(ctx context.Context) dataloader.BlockLoaderConfig {
	if !HasFieldAny(ctx, []string{"size", "tx_count"}) {
		return blockLoaderConfigNoInfo
	}
	return blockLoaderConfigWithInfo
}

var blockLoaderConfigNoInfo = dataloader.BlockLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([][]*model.Block, []error) {
		modelBlocks, errs := blockLoad(keys, false)
		if errs != nil {
			return nil, errs
		}
		return modelBlocks, nil
	},
}

var blockLoaderConfigWithInfo = dataloader.BlockLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([][]*model.Block, []error) {
		modelBlocks, errs := blockLoad(keys, true)
		if errs != nil {
			return nil, errs
		}
		return modelBlocks, nil
	},
}

func blockLoad(keys []string, withInfo bool) ([][]*model.Block, []error) {
	var txHashes = make([][32]byte, len(keys))
	for i := range keys {
		hash, err := chainhash.NewHashFromStr(keys[i])
		if err != nil {
			return nil, []error{jerr.Get("error getting tx hash from string for block loader"+
				"", err)}
		}
		txHashes[i] = *hash
	}
	txBlocks, err := chain.GetTxBlocks(txHashes)
	if err != nil {
		return nil, []error{jerr.Get("error getting blocks for tx for block loader", err)}
	}
	var blockHashes = make([][]byte, len(txBlocks))
	for i := range txBlocks {
		blockHashes[i] = txBlocks[i].BlockHash[:]
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
				Hash: chainhash.Hash(txBlock.BlockHash).String(),
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

func TxLoader(ctx context.Context, txHash string) (*model.Tx, error) {
	var raw string
	if HasFieldAny(ctx, []string{"raw"}) {
		var err error
		if raw, err = dataloader.NewTxRawLoader(txRawLoaderConfig).Load(txHash); err != nil {
			return nil, jerr.Get("error getting tx raw from dataloader for post resolver", err)
		}
	}
	var seen model.Date
	if HasField(ctx, "seen") {
		txSeen, err := dataloader.NewTxSeenLoader(txSeenLoaderConfig).Load(txHash)
		if err != nil {
			return nil, jerr.Get("error getting tx seen for tx loader", err)
		}
		seen = *txSeen
	}
	return &model.Tx{
		Hash: txHash,
		Raw:  raw,
		Seen: seen,
	}, nil
}
