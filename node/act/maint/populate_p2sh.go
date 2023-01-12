package maint

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/act/tx_raw"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/dbi"
)

type PopulateP2sh struct {
	BlocksProcessed int
}

func NewPopulateP2sh() *PopulateP2sh {
	return &PopulateP2sh{}
}

func (p *PopulateP2sh) Populate(startHeight int64) error {
	addressSaver := saver.NewAddress(false, false)
	var maxHeight = startHeight
	for {
		heightBlocks, err := chain.GetHeightBlocksAllLimit(maxHeight, false, client.HugeLimit, false)
		if err != nil {
			return jerr.Get("fatal error getting height blocks all for populate p2sh", err)
		}
		for _, heightBlock := range heightBlocks {
			var startIndex uint32
			var totalTxs int
			for {
				blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
					BlockHash:  heightBlock.BlockHash,
					StartIndex: startIndex,
					Limit:      client.LargeLimit,
				})
				if err != nil {
					return jerr.Get("error getting block txs for populate p2sh", err)
				}
				block, err := chain.GetBlock(heightBlock.BlockHash)
				if err != nil {
					return jerr.Get("error getting block info for populate p2sh", err)
				}
				blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
				if err != nil {
					return jerr.Get("error getting block header for populate p2sh", err)
				}
				var txHashes = make([][32]byte, len(blockTxs))
				for i := range blockTxs {
					if blockTxs[i].Index > startIndex {
						startIndex = blockTxs[i].Index
					}
					txHashes[i] = blockTxs[i].TxHash
				}
				txRaws, err := tx_raw.Get(txHashes)
				if err != nil {
					return jerr.Get("error getting tx raws for populate p2sh", err)
				}
				var msgTxs = make([]*wire.MsgTx, len(txRaws))
				for i := range txRaws {
					msgTxs[i], err = memo.GetMsgFromRaw(txRaws[i].Raw)
					if err != nil {
						return jerr.Get("error getting msg tx for populate p2sh", err)
					}
				}
				dbiBlock := dbi.WireBlockToBlock(memo.GetBlockFromTxs(msgTxs, blockHeader))
				if err := addressSaver.SaveTxs(dbiBlock); err != nil {
					return jerr.Get("error saving txs for populate p2sh", err)
				}
				totalTxs += len(blockTxs)
				if len(blockTxs) < client.LargeLimit {
					jlog.Logf("processed block p2sh: %d %s %s (tx: %d, p2sh: %d, p2pkh: %d)\n",
						heightBlock.Height, chainhash.Hash(heightBlock.BlockHash),
						blockHeader.Timestamp.Format("2006-01-02T15:04:05"), totalTxs,
						addressSaver.P2shCount, addressSaver.P2pkhCount)
					p.BlocksProcessed++
					addressSaver.P2pkhCount = 0
					addressSaver.P2shCount = 0
					break
				}
			}
			if heightBlock.Height > maxHeight {
				maxHeight = heightBlock.Height
			}
		}
		if len(heightBlocks) < client.HugeLimit {
			break
		}
	}
	return nil
}
