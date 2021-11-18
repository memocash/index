package get

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type Block struct {
	Height    int64
	BlockHash []byte
	Block     *wire.BlockHeader
}

func (b *Block) Get() error {
	if b.Height > 0 {
		heightBlock, err := item.GetHeightBlockSingle(b.Height)
		if err != nil {
			return jerr.Get("error getting block height hash from db", err)
		}
		b.BlockHash = heightBlock.BlockHash
	} else if len(b.BlockHash) > 0 {
		blockHashHeight, err := item.GetBlockHeight(b.BlockHash)
		if err != nil {
			return jerr.Get("error getting block hash height from db", err)
		}
		b.Height = blockHashHeight.Height
	} else {
		return jerr.New("block height and hash not set")
	}
	queueBlock, err := item.GetBlock(b.BlockHash)
	if err != nil {
		return jerr.Get("error getting block from db", err)
	}
	b.Block, err = memo.GetBlockHeaderFromRaw(queueBlock.Raw)
	if err != nil {
		return jerr.Get("error getting block from raw", err)
	}
	return nil
}

func NewBlock(blockHash []byte) *Block {
	return &Block{
		BlockHash: blockHash,
	}
}

func NewBlockByHeight(height int64) *Block {
	return &Block{
		Height: height,
	}
}
