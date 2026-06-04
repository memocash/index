package maint

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

// PurgeHeightIndex removes corrupted height-index entries (HeightBlock and
// BlockHeight) at or above a start height. Unlike DeleteBlocks it does NOT
// remove the Block, BlockInfo, or block tx data: those are keyed by block hash
// and are shared with the block's legitimate lower-height mapping, so deleting
// them would destroy real data. This is intended for cleaning up a duplicate
// copy of the chain written at inflated heights (e.g. by a ScanHeaders run that
// resumed from a stale max height).
type PurgeHeightIndex struct {
	Ctx          context.Context
	Verbose      bool
	DryRun       bool
	Verify       bool
	Start        int64
	HeightBlocks int
	Skipped      int
	Duplicates   int
}

func NewPurgeHeightIndex(ctx context.Context, start int64, verbose, dryRun, verify bool) *PurgeHeightIndex {
	return &PurgeHeightIndex{
		Ctx:     ctx,
		Start:   start,
		Verbose: verbose,
		DryRun:  dryRun,
		Verify:  verify,
	}
}

func (p *PurgeHeightIndex) Purge() error {
	var nextHeight = p.Start
	for {
		heightBlocks, err := chain.GetHeightBlocksAllLimit(p.Ctx, nextHeight, client.HugeLimit, false)
		if err != nil {
			return fmt.Errorf("error getting height blocks; %w", err)
		}
		if len(heightBlocks) == 0 {
			break
		}
		var objects []db.Object
		for _, hb := range heightBlocks {
			nextHeight = hb.Height + 1
			if p.Verify {
				duplicate, err := p.isDuplicate(hb)
				if err != nil {
					return fmt.Errorf("error verifying height block at %d; %w", hb.Height, err)
				}
				if !duplicate {
					p.Skipped++
					log.Printf("SKIP height %d: %s — block not recorded at any lower height, leaving as-is",
						hb.Height, chainhash.Hash(hb.BlockHash))
					continue
				}
			}
			if p.Verbose || p.DryRun {
				log.Printf("Height %d: %s", hb.Height, chainhash.Hash(hb.BlockHash))
			}
			p.HeightBlocks++
			objects = append(objects,
				&chain.HeightBlock{Height: hb.Height, BlockHash: hb.BlockHash},
				&chain.BlockHeight{Height: hb.Height, BlockHash: hb.BlockHash},
			)
		}
		if len(objects) > 0 && !p.DryRun {
			if err := db.Remove(objects); err != nil {
				return fmt.Errorf("error removing height index objects; %w", err)
			}
		}
	}
	if err := p.purgeHeightDuplicates(); err != nil {
		return fmt.Errorf("error purging height duplicates; %w", err)
	}
	return nil
}

// isDuplicate reports whether the same block hash is also recorded at a lower
// height. If so, the entry at hb.Height is a duplicate and its height mapping
// can be safely removed without losing the block's only height record.
func (p *PurgeHeightIndex) isDuplicate(hb *chain.HeightBlock) (bool, error) {
	heights, err := p.getBlockHeights(hb.BlockHash)
	if err != nil {
		return false, err
	}
	for _, height := range heights {
		if height < hb.Height {
			return true, nil
		}
	}
	return false, nil
}

func (p *PurgeHeightIndex) getBlockHeights(blockHash [32]byte) ([]int64, error) {
	blockHeights, err := chain.GetBlockHeights(p.Ctx, [][32]byte{blockHash})
	if err != nil {
		return nil, fmt.Errorf("error getting block heights; %w", err)
	}
	var heights []int64
	for _, blockHeight := range blockHeights {
		heights = append(heights, blockHeight.Height)
	}
	return heights, nil
}

func (p *PurgeHeightIndex) purgeHeightDuplicates() error {
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		if p.Start > 0 {
			startUid = jutil.GetInt64DataBig(p.Start)
		}
		for {
			if err := dbClient.GetByPrefix(p.Ctx, db.TopicChainHeightDuplicate, client.Prefix{
				Start: startUid,
			}, client.OptionHugeLimit()); err != nil {
				return fmt.Errorf("error getting height duplicates for shard %d; %w", i, err)
			}
			if len(dbClient.Messages) == 0 {
				break
			}
			var objects []db.Object
			for _, msg := range dbClient.Messages {
				var hd chain.HeightDuplicate
				hd.SetUid(msg.Uid)
				if hd.Height >= p.Start {
					if p.Verbose || p.DryRun {
						log.Printf("Height duplicate at height %d: %s",
							hd.Height, chainhash.Hash(hd.BlockHash))
					}
					p.Duplicates++
					objects = append(objects, &chain.HeightDuplicate{
						Height:    hd.Height,
						BlockHash: hd.BlockHash,
					})
				}
			}
			if len(objects) > 0 && !p.DryRun {
				if err := db.Remove(objects); err != nil {
					return fmt.Errorf("error removing height duplicates; %w", err)
				}
			}
			if len(dbClient.Messages) < client.HugeLimit {
				break
			}
			startUid = dbClient.Messages[len(dbClient.Messages)-1].Uid
		}
	}
	return nil
}
