package process

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/status"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
)

const (
	BlockProcessLimit = client.MediumLimit
)

type HeightBlock struct {
	Height    int64
	BlockHash []byte
}

type Block struct {
	txSave dbi.TxSave
	Status StatusHeight
	Delay  int
	UseRaw bool
	Shard  int
}

// Process goes through all blocks, read tx_outputs, save outputs
func (t *Block) Process() error {
	statusHeight, err := t.Status.GetHeight()
	if err != nil {
		return jerr.Get("error getting blocks status height", err)
	}
	var height = statusHeight.Height
	var waitForBlocks bool
	var delayBlocks []*HeightBlock
	if t.Delay > 0 {
		jlog.Logf("Using delay: %d\n", t.Delay)
	}
	if t.Shard != status.NoShard {
		jlog.Logf("Using shard: %d\n", t.Shard)
	}
	for {
		var heightBlocks []*HeightBlock
		if t.UseRaw {
			heightBlockItems, err := item.GetHeightBlocksAll(height+1, waitForBlocks)
			if err != nil {
				return jerr.Getf(err, "error no block raws returned for block process, height: %d", height)
			}
			for _, heightBlockItem := range heightBlockItems {
				heightBlocks = append(heightBlocks, &HeightBlock{
					Height:    heightBlockItem.Height,
					BlockHash: heightBlockItem.BlockHash,
				})
			}
		} else if t.Shard != status.NoShard {
			heightBlockItems, err := item.GetHeightBlockShardsAll(uint(t.Shard), height+1, waitForBlocks)
			if err != nil {
				return jerr.Getf(err, "error no block shards returned for block process, height: %d", height)
			}
			for _, heightBlockItem := range heightBlockItems {
				heightBlocks = append(heightBlocks, &HeightBlock{
					Height:    heightBlockItem.Height,
					BlockHash: heightBlockItem.BlockHash,
				})
			}
		} else {
			return jerr.New("error not using raw or shard for block tx processor")
		}
		heightBlocks = append(delayBlocks, heightBlocks...)
		delayIndex := len(heightBlocks) - t.Delay
		if delayIndex < 0 {
			delayIndex = 0
		}
		delayBlocks = heightBlocks[delayIndex:]
		var maxHeight int64
		var processedCount int
		for i := range heightBlocks {
			var processBlock = i < delayIndex
			if processBlock {
				if err := t.ProcessBlock(heightBlocks[i]); err != nil {
					return jerr.Get("error processing block height", err)
				}
				processedCount++
			}
			if heightBlocks[i].Height > maxHeight {
				maxHeight = heightBlocks[i].Height
				if processBlock {
					if err := t.Status.SetHeight(status.BlockHeight{
						Height: heightBlocks[i].Height,
						Block:  heightBlocks[i].BlockHash,
					}); err != nil {
						return jerr.Get("error setting block processor status height", err)
					}
				}
			}
		}
		if processedCount > 1 {
			jlog.Logf("Processed block heights: %d, height: %d\n", processedCount, height-int64(t.Delay))
		}
		if maxHeight == 0 || height == maxHeight {
			if !waitForBlocks {
				waitForBlocks = true
				continue
			} else if len(heightBlocks) < client.LargeLimit {
				return nil
			}
			return jerr.Newf("error height = max height, possible loop: %d", height)
		}
		if maxHeight > height {
			height = maxHeight
		}
		waitForBlocks = false
	}
}

func (t *Block) ProcessBlock(heightBlock *HeightBlock) error {
	block, err := item.GetBlock(heightBlock.BlockHash)
	if err != nil {
		return jerr.Getf(err, "error getting block: %d %x", heightBlock.Height,
			jutil.ByteReverse(heightBlock.BlockHash))
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return jerr.Get("error getting block header from raw", err)
	}
	var txCount int
	for _, shard := range config.GetQueueShards() {
		if t.Shard != status.NoShard && int(shard.Shard) != t.Shard {
			continue
		}
		var lastTxHashReverse []byte
		for {
			blockTxesRaw, err := item.GetBlockTxesRaw(item.BlockTxesRawRequest{
				Shard:       shard.Shard,
				BlockHash:   heightBlock.BlockHash,
				StartTxHash: jutil.ByteReverse(lastTxHashReverse),
				Limit:       BlockProcessLimit,
			})
			if err != nil {
				return jerr.Get("error getting block txes", err)
			}
			txCount += len(blockTxesRaw)
			for i := range blockTxesRaw {
				reverseTxHash := jutil.ByteReverse(blockTxesRaw[i].TxHash)
				if bytes.Compare(reverseTxHash, lastTxHashReverse) == 1 {
					lastTxHashReverse = reverseTxHash
				}
			}
			var msgTxs = make([]*wire.MsgTx, len(blockTxesRaw))
			for i := range blockTxesRaw {
				msgTxs[i], err = memo.GetMsgFromRaw(blockTxesRaw[i].Raw)
				if err != nil {
					return jerr.Get("error getting tx from raw block tx", err)
				}
			}
			dbiBlock := dbi.GetBlockWithHeight(memo.GetBlockFromTxs(msgTxs, blockHeader), heightBlock.Height)
			if err = t.txSave.SaveTxs(dbiBlock); err != nil {
				return jerr.Get("error saving block txs", err)
			}
			if len(blockTxesRaw) < BlockProcessLimit {
				break
			}
		}
	}
	jlog.Logf("Block: %s %d %s, txs: %s\n", hs.GetTxString(heightBlock.BlockHash), heightBlock.Height,
		blockHeader.Timestamp.Format("2006-01-02 15:04:05"), jfmt.AddCommasInt(txCount))
	return nil
}

func NewBlockRaw(shard int, statusHeight StatusHeight, txSave dbi.TxSave) *Block {
	return &Block{
		Status: statusHeight,
		txSave: txSave,
		UseRaw: true,
		Shard:  shard,
	}
}

func NewBlockShard(shard int, status StatusHeight, txSave dbi.TxSave) *Block {
	return &Block{
		Status: status,
		txSave: txSave,
		Shard:  shard,
	}
}
