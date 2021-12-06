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
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
)

const (
	BlockProcessLimit = client.LargeLimit
)

type Block struct {
	txSave dbi.TxSave
	Status StatusHeight
}

// Process goes through all blocks, read tx_outputs, save outputs
func (t *Block) Process() error {
	statusHeight, err := t.Status.GetHeight()
	if err != nil {
		return jerr.Get("error getting blocks status height", err)
	}
	var height = statusHeight.Height
	var waitForBlocks bool
	for {
		heightBlocks, err := item.GetHeightBlocksAll(height, waitForBlocks)
		if err != nil {
			return jerr.Getf(err, "error no blocks returned for block process, height: %d", height)
		}
		var maxHeight int64
		for i := range heightBlocks {
			block, err := item.GetBlock(heightBlocks[i].BlockHash)
			if err != nil {
				return jerr.Getf(err, "error getting block: %d %x", heightBlocks[i].Height,
					jutil.ByteReverse(heightBlocks[i].BlockHash))
			}
			blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
			if err != nil {
				return jerr.Get("error getting block header from raw", err)
			}
			var txCount int
			for _, shard := range config.GetQueueShards() {
				var lastTxHashReverse []byte
				for {
					blockTxesRaw, err := item.GetBlockTxesRaw(item.BlockTxesRawRequest{
						Shard:       shard.Min,
						BlockHash:   heightBlocks[i].BlockHash,
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
					err = t.txSave.SaveTxs(memo.GetBlockFromTxs(msgTxs, blockHeader))
					if err != nil {
						return jerr.Get("error saving block txs", err)
					}
					if len(blockTxesRaw) < BlockProcessLimit {
						break
					}
				}
			}
			if heightBlocks[i].Height > maxHeight {
				maxHeight = heightBlocks[i].Height
				err = t.Status.SetHeight(BlockHeight{
					Height: heightBlocks[i].Height,
					Block:  heightBlocks[i].BlockHash,
				})
				if err != nil {
					return jerr.Get("error setting block processor status height", err)
				}
			}
			jlog.Logf("Block: %s %d %s, txs: %s\n", hs.GetTxString(heightBlocks[i].BlockHash), heightBlocks[i].Height,
				blockHeader.Timestamp.Format("2006-01-02 15:04:05"), jfmt.AddCommasInt(txCount))
		}
		if len(heightBlocks) > 1 {
			jlog.Logf("Processed block heights: %d, height: %d\n", len(heightBlocks), height)
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

func NewBlock(status StatusHeight, txSave dbi.TxSave) *Block {
	return &Block{
		Status: status,
		txSave: txSave,
	}
}
