package saver

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"log"
)

type Block struct {
	Verbose         bool
	BlockHash       chainhash.Hash
	PrevBlockHash   chainhash.Hash
	PrevBlockHeight int64
	NewHeight       int64
}

func (b *Block) SaveBlock(info dbi.BlockInfo) error {
	b.BlockHash = info.Header.BlockHash()
	if err := b.saveBlockObjects(info); err != nil {
		return fmt.Errorf("error saving block objects; %w", err)
	}
	return nil
}

func (b *Block) saveBlockObjects(info dbi.BlockInfo) error {
	var objects = make([]db.Object, 1)
	if b.Verbose {
		log.Printf("saving block: %s (parent: %s)\n", info.Header.BlockHash(), info.Header.PrevBlock.String())
	}
	headerRaw := memo.GetRawBlockHeader(info.Header)
	objects[0] = &chain.Block{
		Hash: b.BlockHash,
		Raw:  headerRaw,
	}
	var parentHeight int64
	var hasParent bool
	if info.Header.PrevBlock == b.PrevBlockHash {
		parentHeight = b.PrevBlockHeight
		hasParent = true
	} else {
		parentBlockHeight, err := chain.GetBlockHeight(info.Header.PrevBlock)
		if err != nil && !client.IsEntryNotFoundError(err) {
			return fmt.Errorf("error getting parent block height for potential orphan; %w", err)
		}
		if parentBlockHeight != nil {
			parentHeight = parentBlockHeight.Height
			hasParent = true
			if b.PrevBlockHash != [32]byte{} {
				objects = append(objects, &chain.HeightDuplicate{
					Height:    parentHeight + 1,
					BlockHash: b.BlockHash,
				})
			}
		}
	}
	if hasParent {
		b.NewHeight = parentHeight + 1
	} else {
		initBlockParent, err := chainhash.NewHashFromStr(config.GetInitBlockParent())
		if err != nil {
			return fmt.Errorf("error parsing init block hash; %w", err)
		}
		if *initBlockParent == info.Header.PrevBlock {
			b.NewHeight = int64(config.GetInitBlockHeight())
		} else {
			b.NewHeight = 0
			// block does not match parent or config init block
		}
	}
	var heightBlock *chain.HeightBlock
	if b.NewHeight != 0 {
		heightBlock = &chain.HeightBlock{
			Height:    b.NewHeight,
			BlockHash: b.BlockHash,
		}
		var blockHeight = &chain.BlockHeight{
			Height:    b.NewHeight,
			BlockHash: b.BlockHash,
		}
		objects = append(objects, blockHeight)
		b.PrevBlockHeight = b.NewHeight
		b.PrevBlockHash = b.BlockHash
	}
	if info.Size > 0 {
		objects = append(objects, &chain.BlockInfo{
			BlockHash: b.BlockHash,
			Size:      info.Size,
			TxCount:   info.TxCount,
		})
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving new db block objects; %w", err)
	}
	if heightBlock != nil {
		// Save height block afterward to avoid race conditions with listeners not being able to find block info
		if err := db.Save([]db.Object{heightBlock}); err != nil {
			return fmt.Errorf("error saving height block; %w", err)
		}
	}
	return nil
}

func (b *Block) GetBlock(heightBack int64) (*chainhash.Hash, error) {
	heightBlock, err := chain.GetRecentHeightBlock()
	if err != nil {
		return nil, fmt.Errorf("error getting recent height block from queue; %w", err)
	}
	if heightBlock == nil {
		return nil, nil
	}
	if heightBack > 0 {
		height := heightBlock.Height - heightBack
		heightBlock, err = chain.GetHeightBlockSingle(height)
		if err != nil {
			return nil, fmt.Errorf("error getting height back height block (height: %d, back: %d); %w",
				height, heightBack, err)
		}
	}
	b.PrevBlockHash = heightBlock.BlockHash
	b.PrevBlockHeight = heightBlock.Height
	blockHash := chainhash.Hash(heightBlock.BlockHash)
	return &blockHash, nil
}

func NewBlock(verbose bool) *Block {
	return &Block{
		Verbose: verbose,
	}
}
