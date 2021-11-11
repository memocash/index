package resolver

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/admin/graph/dataloader"
	"github.com/memocash/server/admin/graph/model"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
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

var blockLoaderConfig = dataloader.BlockLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []string) ([][]*model.Block, []error) {
		var txHashes = make([][]byte, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i])
			if err != nil {
				return nil, []error{jerr.Get("error getting tx hash from string for block resolver", err)}
			}
			txHashes[i] = hash.CloneBytes()
		}
		txBlocks, err := item.GetTxBlocks(txHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting blocks for tx for resolver", err)}
		}
		var blockHashes = make([][]byte, len(txBlocks))
		for i := range txBlocks {
			blockHashes[i] = txBlocks[i].BlockHash
		}
		blocks, err := item.GetBlocks(blockHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting blocks for tx resolver", err)}
		}
		blockHeights, err := item.GetBlockHeights(blockHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting block heights for tx resolver", err)}
		}
		var modelBlocks = make([][]*model.Block, len(txHashes))
		for i := range txHashes {
			for _, txBlock := range txBlocks {
				if !bytes.Equal(txBlock.TxHash, txHashes[i]) {
					continue
				}
				var modelBlock = &model.Block{
					Hash: hs.GetTxString(txBlock.BlockHash),
				}
				for _, block := range blocks {
					if !bytes.Equal(block.Hash, txBlock.BlockHash) {
						continue
					}
					rawBlock, err := memo.GetBlockFromRaw(block.Raw)
					if err != nil {
						return nil, []error{jerr.Get("error getting block from raw for tx resolver", err)}
					}
					modelBlock.Timestamp = model.Date(rawBlock.Header.Timestamp)
					for _, blockHeight := range blockHeights {
						if bytes.Equal(blockHeight.BlockHash, block.Hash) {
							height := int(blockHeight.Height)
							modelBlock.Height = &height
						}
					}
				}
				modelBlocks[i] = append(modelBlocks[i], modelBlock)
			}
		}
		return modelBlocks, nil
	},
}
