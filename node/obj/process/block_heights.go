package process

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/dbi"
)

type BlockHeights struct {
	Shard    uint32
	Topic    string
	dbiBlock dbi.BlockHeightSave
	Status   Status
}

// Read height_blocks and save corresponding block_heights
func (t *BlockHeights) Process() error {
	height, err := t.Status.GetHeight()
	if err != nil {
		return jerr.Get("error getting blocks status height", err)
	}
	for {
		heightBlocks, err := item.GetHeightBlocks(t.Shard, height, false)
		if len(heightBlocks) == 0 {
			return jerr.Newf("error no blocks returned for obj process, height: %d", height)
		}
		var maxHeight int64
		var dbiBlockHeights = make([]*dbi.BlockHeight, len(heightBlocks))
		for i := range heightBlocks {
			jlog.Logf("Height: %d, block hash: %s\n", heightBlocks[i].Height, hs.GetTxString(heightBlocks[i].BlockHash))
			if heightBlocks[i].Height > maxHeight {
				maxHeight = heightBlocks[i].Height
			}
			dbiBlockHeights[i] = &dbi.BlockHeight{
				BlockHash: heightBlocks[i].BlockHash,
				Height:    heightBlocks[i].Height,
			}
		}
		err = t.dbiBlock.SaveHeights(dbiBlockHeights)
		if err != nil {
			return jerr.Get("error saving block heights", err)
		}
		jlog.Logf("Processed block heights: %d, max height: %d\n",
			len(heightBlocks), maxHeight)
		if height == maxHeight {
			return jerr.Newf("error height = max height, possible loop: %d", height)
		}
		err = t.Status.SetHeight(maxHeight)
		if err != nil {
			return jerr.Get("error setting block txs processor status height", err)
		}
		height = maxHeight
	}
}

func NewBlockHeights(shard uint32, status Status, dbiBlock dbi.BlockHeightSave) *BlockHeights {
	return &BlockHeights{
		Shard:    shard,
		Status:   status,
		dbiBlock: dbiBlock,
	}
}
