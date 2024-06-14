package sub

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	"log"
)

type Block struct {
	Name         string
	BlockHashChan chan [32]byte
	Cancel       context.CancelFunc
}

func (r *Block) Listen(ctx context.Context) (<-chan *model.Block, error) {
	ctx, r.Cancel = context.WithCancel(ctx)
	var blockChan = make(chan *model.Block)
	r.BlockHashChan = make(chan [32]byte)
	blockHeightListener, err := chain.ListenBlockHeights(ctx)
	if err != nil {
		r.Cancel()
		return nil, fmt.Errorf("error getting block height listener for subscription; %w", err)
	}
	go func() {
		defer func() {
			close(r.BlockHashChan)
			close(blockChan)
			r.Cancel()
		}()
		for {
			var blockHeight *chain.BlockHeight
			var ok bool
			select {
			case <-ctx.Done():
				return
			case blockHeight, ok = <-blockHeightListener:
				if !ok {
					return
				}
			}
			var block = &model.Block{
				Hash:   blockHeight.BlockHash,
				Height: model.IntPtr(int(blockHeight.Height)),
			}
			if err := attach.ToBlocks(ctx, attach.GetFields(ctx), []*model.Block{block}); err != nil {
				log.Printf("error attaching to blocks for subscription; %v", err)
				return
			}
			blockChan <- block
		}
	}()
	return blockChan, nil
}
