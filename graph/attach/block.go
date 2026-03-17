package attach

import (
	"context"
	"fmt"

	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Blocks struct {
	base
	Blocks []*model.Block
}

func ToBlocks(ctx context.Context, fields []Field, blocks []*model.Block) error {
	if len(blocks) == 0 {
		return nil
	}
	b := Blocks{
		base:   base{Ctx: ctx, Fields: fields},
		Blocks: blocks,
	}
	b.Wait.Add(4)
	go b.AttachRaws()
	go b.AttachHeights()
	go b.AttachInfos()
	go b.AttachTxs()
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

func (b *Blocks) getBlockIndexMap() map[[32]byte][]int {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	m := make(map[[32]byte][]int, len(b.Blocks))
	for i := range b.Blocks {
		m[b.Blocks[i].Hash] = append(m[b.Blocks[i].Hash], i)
	}
	return m
}

func (b *Blocks) AttachRaws() {
	defer b.Wait.Done()
	if !b.HasField([]string{"raw", "timestamp"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blocks, err := chain.GetBlocks(b.Ctx, blockHashes)
	if err != nil {
		b.AddError(fmt.Errorf("error getting blocks for block loader; %w", err))
		return
	}
	blockIndexMap := b.getBlockIndexMap()
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for j := range blocks {
		indices, ok := blockIndexMap[blocks[j].Hash]
		if !ok {
			continue
		}
		blockHeader, err := memo.GetBlockHeaderFromRaw(blocks[j].Raw)
		if err != nil {
			b.AddError(fmt.Errorf("error getting block header from raw for block loader; %w", err))
			return
		}
		for _, i := range indices {
			b.Blocks[i].Timestamp = model.Date(blockHeader.Timestamp)
			b.Blocks[i].Raw = blocks[j].Raw
		}
	}
}

func (b *Blocks) AttachHeights() {
	defer b.Wait.Done()
	if !b.HasField([]string{"height"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blockHeights, err := chain.GetBlockHeights(b.Ctx, blockHashes)
	if err != nil {
		b.AddError(fmt.Errorf("error getting block heights for block loader; %w", err))
		return
	}
	blockIndexMap := b.getBlockIndexMap()
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for j := range blockHeights {
		indices, ok := blockIndexMap[blockHeights[j].BlockHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			b.Blocks[i].Height = model.IntPtr(int(blockHeights[j].Height))
		}
	}
}

func (b *Blocks) AttachInfos() {
	defer b.Wait.Done()
	if !b.HasField([]string{"size", "tx_count"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blockInfos, err := chain.GetBlockInfos(b.Ctx, blockHashes)
	if err != nil {
		b.AddError(fmt.Errorf("error getting block infos for block loader; %w", err))
		return
	}
	blockIndexMap := b.getBlockIndexMap()
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for j := range blockInfos {
		indices, ok := blockIndexMap[blockInfos[j].BlockHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			b.Blocks[i].Size = blockInfos[j].Size
			b.Blocks[i].TxCount = blockInfos[j].TxCount
		}
	}
}

func (b *Blocks) AttachTxs() {
	defer b.Wait.Done()
	if !b.HasField([]string{"txs"}) {
		return
	}
	// TODO: More efficient mutex usage, doesn't need to lock the whole time
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	txsField := b.Fields.GetField("txs")
	var startIndex uint32
	if startNum, err := model.UnmarshalUint32(txsField.Arguments["start"]); err == nil {
		startIndex = uint32(startNum)
	}
	var allTxs []*model.Tx
	for _, block := range b.Blocks {
		blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
			Context:    b.Ctx,
			BlockHash:  block.Hash,
			StartIndex: startIndex,
			Limit:      client.DefaultLimit,
		})
		if err != nil {
			b.AddError(fmt.Errorf("error getting block transactions for attach; %w", err))
			return
		}
		block.Txs = make([]*model.TxBlock, len(blockTxs))
		for i := range blockTxs {
			block.Txs[i] = &model.TxBlock{
				Index:  blockTxs[i].Index,
				TxHash: blockTxs[i].TxHash,
				Tx:     &model.Tx{Hash: blockTxs[i].TxHash},
			}
			allTxs = append(allTxs, block.Txs[i].Tx)
		}
	}
	if err := ToTxs(b.Ctx, GetPrefixFields(txsField.Fields, "tx."), allTxs); err != nil {
		b.AddError(fmt.Errorf("error attaching to block transactions; %w", err))
		return
	}
}
