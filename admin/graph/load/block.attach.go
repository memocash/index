package load

import (
	"fmt"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Blocks struct {
	baseA
	Blocks []*model.Block
}

func AttachToBlocks(preloads []string, outputs []*model.Block) error {
	if len(outputs) == 0 {
		return nil
	}
	b := Blocks{
		baseA:  baseA{Preloads: preloads},
		Blocks: outputs,
	}
	b.Wait.Add(3)
	go b.AttachRaws()
	go b.AttachHeights()
	go b.AttachInfos()
	b.Wait.Wait()
	if len(b.Errors) > 0 {
		return fmt.Errorf("error attaching to blocks; %w", b.Errors[0])
	}
	return nil
}

func (b *Blocks) GetBlockHashes() [][32]byte {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	var blockHashes [][32]byte
	for i := range b.Blocks {
		blockHashes = append(blockHashes, b.Blocks[i].Hash)
	}
	return blockHashes
}

func (b *Blocks) AttachRaws() {
	defer b.Wait.Done()
	if !b.HasPreload([]string{"raw", "timestamp"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blocks, err := chain.GetBlocks(blockHashes)
	if err != nil {
		b.AddError(fmt.Errorf("error getting blocks for block loader; %w", err))
		return
	}
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for i := range b.Blocks {
		for j := range blocks {
			if b.Blocks[i].Hash != blocks[j].Hash {
				continue
			}
			blockHeader, err := memo.GetBlockHeaderFromRaw(blocks[j].Raw)
			if err != nil {
				b.AddError(fmt.Errorf("error getting block header from raw for block loader; %w", err))
				return
			}
			b.Blocks[i].Timestamp = model.Date(blockHeader.Timestamp)
			b.Blocks[i].Raw = blocks[j].Raw
			break
		}
	}
}

func (b *Blocks) AttachHeights() {
	defer b.Wait.Done()
	if !b.HasPreload([]string{"height"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blockHeights, err := chain.GetBlockHeights(blockHashes)
	if err != nil {
		b.AddError(fmt.Errorf("error getting block heights for block loader; %w", err))
		return
	}
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for i := range b.Blocks {
		for j := range blockHeights {
			if b.Blocks[i].Hash != blockHeights[j].BlockHash {
				continue
			}
			height := int(blockHeights[j].Height)
			b.Blocks[i].Height = &height
			break
		}
	}
}

func (b *Blocks) AttachInfos() {
	defer b.Wait.Done()
	if !b.HasPreload([]string{"size", "tx_count"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blockInfos, err := chain.GetBlockInfos(blockHashes)
	if err != nil {
		b.AddError(fmt.Errorf("error getting block infos for block loader; %w", err))
		return
	}
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for i := range b.Blocks {
		for j := range blockInfos {
			if b.Blocks[i].Hash != blockInfos[j].BlockHash {
				continue
			}
			b.Blocks[i].Size = blockInfos[j].Size
			b.Blocks[i].TxCount = blockInfos[j].TxCount
			break
		}
	}
}
