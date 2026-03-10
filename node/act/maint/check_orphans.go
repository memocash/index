package maint

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
)

type CheckOrphans struct {
	Ctx            context.Context
	Verbose        bool
	Save           bool
	Total          int
	Orphans        int
	Breaks         int
	FalsePositives int
	Saved          int
}

func NewCheckOrphans(ctx context.Context, verbose bool, save bool) *CheckOrphans {
	return &CheckOrphans{
		Ctx:     ctx,
		Verbose: verbose,
		Save:    save,
	}
}

func (c *CheckOrphans) Check() error {
	var nextHeight int64 = 1
	var prevHB *chain.HeightBlock
	for {
		heightBlocks, err := chain.GetHeightBlocksAllLimit(c.Ctx, nextHeight, client.HugeLimit, false)
		if err != nil {
			return fmt.Errorf("error getting height blocks; %w", err)
		}
		if len(heightBlocks) == 0 {
			break
		}
		for _, hb := range heightBlocks {
			c.Total++
			if prevHB != nil && hb.Height != nextHeight {
				if hb.Height == prevHB.Height {
					c.Orphans++
					log.Printf("Height duplicate at %d:\n - %s\n - %s\n",
						hb.Height, chainhash.Hash(hb.BlockHash), chainhash.Hash(prevHB.BlockHash))
					if c.Save {
						if err := db.Save([]db.Object{
							&chain.HeightDuplicate{Height: hb.Height, BlockHash: hb.BlockHash},
							&chain.HeightDuplicate{Height: prevHB.Height, BlockHash: prevHB.BlockHash},
						}); err != nil {
							return fmt.Errorf("error saving height duplicates at %d; %w", hb.Height, err)
						}
						c.Saved += 2
					}
				} else {
					// Gap detected — verify if missing heights actually exist
					for missingHeight := nextHeight; missingHeight < hb.Height; missingHeight++ {
						heightBlocksAtHeight, err := chain.GetHeightBlock(c.Ctx, missingHeight)
						if err != nil {
							log.Printf("Error checking height %d: %v", missingHeight, err)
							c.Breaks++
							continue
						}
						if len(heightBlocksAtHeight) == 0 {
							c.Breaks++
							log.Printf("Missing height %d (between %d and %d)",
								missingHeight, prevHB.Height, hb.Height)
						} else {
							c.FalsePositives++
							if c.Verbose {
								log.Printf("False positive gap at height %d (exists on shard, %d blocks)",
									missingHeight, len(heightBlocksAtHeight))
							}
						}
					}
				}
			}
			nextHeight = hb.Height + 1
			prevHB = hb
		}
	}
	return nil
}
