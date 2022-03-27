package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type Block struct {
	Verbose         bool
	BlockHash       chainhash.Hash
	BlockHashBytes  []byte
	PrevBlockHash   []byte
	PrevBlockHeight int64
}

func (t *Block) SaveBlock(header wire.BlockHeader) error {
	t.BlockHash = header.BlockHash()
	t.BlockHashBytes = t.BlockHash.CloneBytes()
	err := t.saveBlockObjects(header)
	if err != nil {
		return jerr.Get("error saving block objects", err)
	}
	return nil
}

func (t *Block) saveBlockObjects(header wire.BlockHeader) error {
	var objects = make([]item.Object, 1)
	if t.Verbose {
		jlog.Logf("block: %s\n", t.BlockHash.String())
	}
	headerRaw := memo.GetRawBlockHeader(header)
	objects[0] = &item.Block{
		Hash: t.BlockHashBytes,
		Raw:  headerRaw,
	}
	var parentHashBytes = header.PrevBlock.CloneBytes()
	var parentHeight int64
	var hasParent bool
	if bytes.Equal(parentHashBytes, t.PrevBlockHash) {
		parentHeight = t.PrevBlockHeight
		hasParent = true
	} else {
		parentBlockHeight, err := item.GetBlockHeight(header.PrevBlock.CloneBytes())
		if err != nil && !client.IsEntryNotFoundError(err) {
			return jerr.Get("error getting parent block height for potential orphan", err)
		}
		if parentBlockHeight != nil {
			parentHeight = parentBlockHeight.Height
			hasParent = true
			if len(t.PrevBlockHash) > 0 {
				objects = append(objects, &item.HeightDuplicate{
					Height:    parentHeight + 1,
					BlockHash: t.BlockHashBytes,
				})
			}
		}
	}
	var skipHeight bool
	var newBlockHeight int64
	if hasParent {
		newBlockHeight = parentHeight + 1
	} else {
		initBlockParent, err := chainhash.NewHashFromStr(config.GetInitBlockParent())
		if err != nil {
			return jerr.Get("error parsing init block hash", err)
		}
		if bytes.Equal(initBlockParent.CloneBytes(), parentHashBytes) {
			newBlockHeight = int64(config.GetInitBlockHeight())
		} else {
			skipHeight = true
			// block does not match parent or config init block
		}
	}
	var heightBlockRaw *item.HeightBlockRaw
	if !skipHeight {
		heightBlockRaw = &item.HeightBlockRaw{
			Height:    newBlockHeight,
			BlockHash: t.BlockHashBytes,
		}
		var blockHeight = &item.BlockHeight{
			Height:    newBlockHeight,
			BlockHash: t.BlockHashBytes,
		}
		objects = append(objects, blockHeight)
		t.PrevBlockHeight = newBlockHeight
		t.PrevBlockHash = t.BlockHashBytes
	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving new db block objects", err)
	}
	if heightBlockRaw != nil {
		// Save height block afterward to avoid race conditions with listeners not being able to find block info
		if err := item.Save([]item.Object{heightBlockRaw}); err != nil {
			return jerr.Get("error saving height block raw", err)
		}
	}
	return nil
}

func (t *Block) GetBlock(heightBack int64) ([]byte, error) {
	heightBlockRaw, err := item.GetRecentHeightBlockRaw()
	if err != nil {
		return nil, jerr.Get("error getting recent height block raw from queue", err)
	}
	if heightBlockRaw == nil {
		return nil, nil
	}
	if heightBack > 0 {
		height := heightBlockRaw.Height - heightBack
		heightBlockRaw, err = item.GetHeightBlockRawSingle(height)
		if err != nil {
			return nil, jerr.Getf(err, "error getting height back height block raw (height: %d, back: %d)",
				height, heightBack)
		}
	}
	t.PrevBlockHash = heightBlockRaw.BlockHash
	t.PrevBlockHeight = heightBlockRaw.Height
	return heightBlockRaw.BlockHash, nil
}

func NewBlock(verbose bool) *Block {
	return &Block{
		Verbose: verbose,
	}
}
