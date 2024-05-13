package load

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Blocks struct {
	baseA
	Blocks []*model.Block
}

func AttachToBlocks(ctx context.Context, fields []Field, blocks []*model.Block) error {
	if len(blocks) == 0 {
		return nil
	}
	b := Blocks{
		baseA:  baseA{Ctx: ctx, Fields: fields},
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
	if !b.HasField([]string{"height"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blockHeights, err := chain.GetBlockHeights(b.Ctx, blockHashes)
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
	if !b.HasField([]string{"size", "tx_count"}) {
		return
	}
	blockHashes := b.GetBlockHashes()
	blockInfos, err := chain.GetBlockInfos(b.Ctx, blockHashes)
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
	if startNum, ok := txsField.Arguments["start"].(json.Number); ok {
		start64, _ := startNum.Int64()
		startIndex = uint32(start64)
	}
	var allTxs []*model.Tx
	for _, block := range b.Blocks {
		blockTxs, err := chain.GetBlockTxs(chain.BlockTxsRequest{
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
	prefixFields := GetPrefixFields(txsField.Fields, "tx.")
	if err := AttachToTxs(b.Ctx, prefixFields, allTxs); err != nil {
		b.AddError(fmt.Errorf("error attaching to block transactions; %w", err))
		return
	}
}
