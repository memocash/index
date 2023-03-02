package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
)

// Tx is the resolver for the tx field.
func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	tx, err := TxLoader(ctx, hash)
	if err != nil {
		return nil, jerr.Get("error getting tx from dataloader for tx query resolver", err)
	}
	return tx, nil
}

// Txs is the resolver for the txs field.
func (r *queryResolver) Txs(ctx context.Context, hashes []string) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// Address is the resolver for the address field.
func (r *queryResolver) Address(ctx context.Context, address string) (*model.Lock, error) {
	if HasField(ctx, "balance") {
		// TODO: Reimplement if needed
		return nil, jerr.New("error balance no longer implemented")
	}
	return &model.Lock{
		Address: address,
	}, nil
}

// Addresses is the resolver for the addresses field.
func (r *queryResolver) Addresses(ctx context.Context, addresses []string) ([]*model.Lock, error) {
	if HasField(ctx, "balance") {
		// TODO: Reimplement if needed
		return nil, jerr.New("error balance no longer implemented")
	}
	var locks []*model.Lock
	for _, address := range addresses {
		locks = append(locks, &model.Lock{
			Address: address,
		})
	}
	return locks, nil
}

// Block is the resolver for the block field.
func (r *queryResolver) Block(ctx context.Context, hash string) (*model.Block, error) {
	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, jerr.Get("error parsing block hash for block query resolver", err)
	}
	blockHeight, err := chain.GetBlockHeight(*blockHash)
	if err != nil {
		return nil, jerr.Get("error getting block height for query resolver", err)
	}
	block, err := chain.GetBlock(*blockHash)
	if err != nil {
		return nil, jerr.Get("error getting raw block", err)
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return nil, jerr.Get("error getting block header from raw", err)
	}
	height := int(blockHeight.Height)
	var modelBlock = &model.Block{
		Hash:      chainhash.Hash(blockHeight.BlockHash).String(),
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
		Raw:       hex.EncodeToString(block.Raw),
	}
	if !HasFieldAny(ctx, []string{"size", "tx_count"}) {
		return modelBlock, nil
	}
	blockInfo, err := chain.GetBlockInfo(*blockHash)
	if err != nil && !client.IsMessageNotSetError(err) {
		return nil, jerr.Get("error getting block infos for query resolver", err)
	}
	if blockInfo != nil {
		modelBlock.Size = blockInfo.Size
		modelBlock.TxCount = blockInfo.TxCount
	}
	return modelBlock, nil
}

// BlockNewest is the resolver for the block_newest field.
func (r *queryResolver) BlockNewest(ctx context.Context) (*model.Block, error) {
	heightBlock, err := chain.GetRecentHeightBlock()
	if err != nil {
		return nil, jerr.Get("error getting recent height block for query", err)
	}
	if heightBlock == nil {
		return nil, nil
	}
	block, err := chain.GetBlock(heightBlock.BlockHash)
	if err != nil {
		return nil, jerr.Get("error getting raw block", err)
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return nil, jerr.Get("error getting block header from raw", err)
	}
	height := int(heightBlock.Height)
	return &model.Block{
		Hash:      chainhash.Hash(heightBlock.BlockHash).String(),
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
	}, nil
}

// Blocks is the resolver for the blocks field.
func (r *queryResolver) Blocks(ctx context.Context, newest *bool, start *uint32) ([]*model.Block, error) {
	var startInt int64
	if start != nil {
		startInt = int64(*start)
	}
	var newestBool bool
	if newest != nil {
		newestBool = *newest
	}
	heightBlocks, err := chain.GetHeightBlocksAllDefault(startInt, false, newestBool)
	if err != nil {
		return nil, jerr.Get("error getting height blocks for query", err)
	}
	var blockHashes = make([][]byte, len(heightBlocks))
	for i := range heightBlocks {
		blockHashes[i] = heightBlocks[i].BlockHash[:]
	}
	blocks, err := chain.GetBlocks(blockHashes)
	if err != nil {
		return nil, jerr.Get("error getting raw blocks", err)
	}
	var blockInfos []*chain.BlockInfo
	if HasFieldAny(ctx, []string{"size", "tx_count"}) {
		if blockInfos, err = chain.GetBlockInfos(blockHashes); err != nil {
			return nil, jerr.Get("error getting block infos for blocks query resolver", err)
		}
	}
	var modelBlocks = make([]*model.Block, len(heightBlocks))
	for i := range heightBlocks {
		var height = int(heightBlocks[i].Height)
		modelBlocks[i] = &model.Block{
			Hash:   chainhash.Hash(heightBlocks[i].BlockHash).String(),
			Height: &height,
		}
		for _, block := range blocks {
			if block.Hash == heightBlocks[i].BlockHash {
				blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
				if err != nil {
					return nil, jerr.Get("error getting block header from raw", err)
				}
				modelBlocks[i].Timestamp = model.Date(blockHeader.Timestamp)
			}
		}
		for _, blockInfo := range blockInfos {
			if blockInfo.BlockHash == heightBlocks[i].BlockHash {
				modelBlocks[i].Size = blockInfo.Size
				modelBlocks[i].TxCount = blockInfo.TxCount
			}
		}
	}
	return modelBlocks, nil
}

// Profiles is the resolver for the profiles field.
func (r *queryResolver) Profiles(ctx context.Context, addresses []string) ([]*model.Profile, error) {
	var profiles []*model.Profile
	for _, addressString := range addresses {
		profile, err := dataloader.NewProfileLoader(load.ProfileLoaderConfig).Load(addressString)
		if err != nil {
			return nil, jerr.Get("error getting profile from dataloader for profile query resolver", err)
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context, txHashes []string) ([]*model.Post, error) {
	posts, errs := dataloader.NewPostLoader(load.PostLoaderConfig).LoadAll(txHashes)
	for i, err := range errs {
		if err != nil {
			return nil, jerr.Getf(err, "error getting post from post dataloader for query resolver: %s", txHashes[i])
		}
	}
	return posts, nil
}

// Room is the resolver for the room field.
func (r *queryResolver) Room(ctx context.Context, name string) (*model.Room, error) {
	return &model.Room{Name: name}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
