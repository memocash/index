package maint

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
)

type CheckOrphans struct {
	Ctx     context.Context
	Verbose bool
	Total   int
	Orphans int
	Breaks  int
}

func NewCheckOrphans(ctx context.Context, verbose bool) *CheckOrphans {
	return &CheckOrphans{
		Ctx:     ctx,
		Verbose: verbose,
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
			if prevHB != nil && hb.Height != nextHeight {
				log.Printf("Found duplicate:\n - %d: %s\n - %d: %s\n",
					hb.Height, chainhash.Hash(hb.BlockHash),
					prevHB.Height, chainhash.Hash(prevHB.BlockHash))
			}
			nextHeight = hb.Height + 1
			prevHB = hb
		}
	}
	return nil
}
