package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/act/tx_raw"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/dbi"
	"log"
)

type PopulateP2sh struct {
	BlocksProcessed int
	Ctx             context.Context
}

func NewPopulateP2sh(ctx context.Context) *PopulateP2sh {
	return &PopulateP2sh{
		Ctx: ctx,
	}
}

func (p *PopulateP2sh) Populate(startHeight int64) error {
	addressSaver := saver.NewAddress(false)
	addressSaver.SkipP2pkh = true
	var maxHeight = startHeight
	for {
		heightBlocks, err := chain.GetHeightBlocksAllLimit(p.Ctx, maxHeight, client.HugeLimit, false)
		if err != nil {
			return fmt.Errorf("fatal error getting height blocks all for populate p2sh; %w", err)
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
					return fmt.Errorf("error getting block txs for populate p2sh; %w", err)
				}
				block, err := chain.GetBlock(p.Ctx, heightBlock.BlockHash)
				if err != nil {
					return fmt.Errorf("error getting block info for populate p2sh; %w", err)
				}
				blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
				if err != nil {
					return fmt.Errorf("error getting block header for populate p2sh; %w", err)
				}
				var txHashShards = make(map[uint32][][32]byte, len(blockTxs))
				for i := range blockTxs {
					if blockTxs[i].Index > startIndex {
						startIndex = blockTxs[i].Index
					}
					shard := db.GetShardIdFromByte32(blockTxs[i].TxHash[:])
					txHashShards[shard] = append(txHashShards[shard], blockTxs[i].TxHash)
				}
				var shardProcess ShardProcess
				shardProcess.Wg.Add(len(txHashShards))
				for shardT, txHashesT := range txHashShards {
					go func(shard uint32, txHashes [][32]byte) {
						defer shardProcess.Wg.Done()
						txRaws, err := tx_raw.Get(p.Ctx, txHashes)
						if err != nil {
							shardProcess.AddError(shard, fmt.Errorf("error getting tx raws for populate p2sh; %w", err))
							return
						}
						var msgTxs = make([]*wire.MsgTx, len(txRaws))
						for i := range txRaws {
							msgTxs[i], err = memo.GetMsgFromRaw(txRaws[i].Raw)
							if err != nil {
								shardProcess.AddError(shard, fmt.Errorf("error getting msg tx for populate p2sh; %w", err))
								return
							}
						}
						dbiBlock := dbi.WireBlockToBlock(memo.GetBlockFromTxs(msgTxs, blockHeader))
						if err := addressSaver.SaveTxs(p.Ctx, dbiBlock); err != nil {
							shardProcess.AddError(shard, fmt.Errorf("error saving txs for populate p2sh; %w", err))
							return
						}
					}(shardT, txHashesT)
				}
				shardProcess.Wg.Wait()
				if len(shardProcess.Errors) > 0 {
					return fmt.Errorf("error processing tx group for populate p2sh; %w", jerr.Combine(shardProcess.Errors...))
				}
				totalTxs += len(blockTxs)
				if len(blockTxs) < client.LargeLimit {
					log.Printf("processed block p2sh: %d %s %s (tx: %d, p2sh: %d, p2pkh: %d)\n",
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
