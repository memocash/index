package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/admin/graph/sub"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Tx is the resolver for the tx field.
func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	txHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, jerr.Get("error getting tx hash from hash", err)
	}
	txHashString := txHash.String()
	preloads := GetPreloads(ctx)
	var raw string
	if jutil.StringsInSlice([]string{"raw", "inputs", "outputs"}, preloads) {
		if raw, err = dataloader.NewTxRawLoader(txRawLoaderConfig).Load(txHashString); err != nil {
			return nil, jerr.Get("error getting tx raw from dataloader for tx query resolver", err)
		}
	}
	return &model.Tx{
		Hash: txHashString,
		Raw:  raw,
	}, nil
}

// Txs is the resolver for the txs field.
func (r *queryResolver) Txs(ctx context.Context, hashes []string) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// Address is the resolver for the address field.
func (r *queryResolver) Address(ctx context.Context, address string) (*model.Lock, error) {
	balance, err := get.NewBalanceFromAddress(address)
	if err != nil {
		return nil, jerr.Get("error getting address from string for balance", err)
	}
	if err := balance.GetBalanceByUtxos(); err != nil {
		return nil, jerr.Get("error getting address balance from network", err)
	}
	return &model.Lock{
		Address: balance.Address,
		Balance: balance.Balance,
	}, nil
}

// Addresses is the resolver for the addresses field.
func (r *queryResolver) Addresses(ctx context.Context, addresses []string) ([]*model.Lock, error) {
	var locks []*model.Lock
	for _, address := range addresses {
		balance, err := get.NewBalanceFromAddress(address)
		if err != nil {
			return nil, jerr.Get("error getting address from string for balances", err)
		}
		if err := balance.GetBalance(); err != nil {
			return nil, jerr.Get("error getting address balance from network (multi)", err)
		}
		locks = append(locks, &model.Lock{
			Address: balance.Address,
			Balance: balance.Balance,
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
	blockHeight, err := chain.GetBlockHeight(blockHash[:])
	if err != nil {
		return nil, jerr.Get("error getting block height for query resolver", err)
	}
	block, err := chain.GetBlock(blockHash[:])
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
	preloads := GetPreloads(ctx)
	if !jutil.StringsInSlice([]string{"size", "tx_count"}, preloads) {
		return modelBlock, nil
	}
	blockInfo, err := chain.GetBlockInfo(blockHash[:])
	if err != nil {
		return nil, jerr.Get("error getting block infos for query resolver", err)
	}
	modelBlock.Size = blockInfo.Size
	modelBlock.TxCount = blockInfo.TxCount
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
	block, err := chain.GetBlock(heightBlock.BlockHash[:])
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
	if jutil.StringsInSlice([]string{"size", "tx_count"}, GetPreloads(ctx)) {
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

// DoubleSpends is the resolver for the double_spends field.
func (r *queryResolver) DoubleSpends(ctx context.Context, newest *bool, start *model.Date) ([]*model.DoubleSpend, error) {
	var startTime time.Time
	if start != nil {
		startTime = time.Time(*start)
	}
	var newestBool bool
	if newest != nil {
		newestBool = *newest
	}
	doubleSpendSeens, err := item.GetDoubleSpendSeensAllLimit(startTime, client.DefaultLimit, newestBool)
	if err != nil {
		return nil, jerr.Get("error getting double spend outputs", err)
	}
	var modelDoubleSpends = make([]*model.DoubleSpend, len(doubleSpendSeens))
	for i := range doubleSpendSeens {
		modelDoubleSpends[i] = &model.DoubleSpend{
			Hash:      hs.GetTxString(doubleSpendSeens[i].TxHash),
			Index:     doubleSpendSeens[i].Index,
			Timestamp: model.Date(doubleSpendSeens[i].Timestamp),
		}
	}
	return modelDoubleSpends, nil
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

// Address is the resolver for the address field.
func (r *subscriptionResolver) Address(ctx context.Context, address string) (<-chan *model.Tx, error) {
	txChan, err := r.Addresses(ctx, []string{address})
	if err != nil {
		return nil, jerr.Get("error getting address for address subscription", err)
	}
	return txChan, nil
}

// Addresses is the resolver for the address field.
func (r *subscriptionResolver) Addresses(ctx context.Context, addresses []string) (<-chan *model.Tx, error) {
	addrs := make([][25]byte, len(addresses))
	for i := range addresses {
		walletAddr, err := wallet.GetAddrFromString(addresses[i])
		if err != nil {
			return nil, jerr.Get("error getting addr for address subscription", err)
		}
		addrs[i] = *walletAddr
	}
	ctx, cancel := context.WithCancel(ctx)
	addrHeightOutputsListeners, err := addr.ListenMempoolAddrHeightOutputsMultiple(ctx, addrs)
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting addr height outputs listener for address subscription", err)
	}
	addrHeightInputsListeners, err := addr.ListenMempoolAddrHeightInputsMultiple(ctx, addrs)
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting addr height inputs listener for address subscription", err)
	}
	var aggregateOutput = make(chan *addr.HeightOutput)
	var aggregateInput = make(chan *addr.HeightInput)
	for _, ch := range addrHeightInputsListeners {
		go func(c chan *addr.HeightInput) {
			for msg := range c {
				aggregateInput <- msg
			}
		}(ch)
	}
	for _, ch := range addrHeightOutputsListeners {
		go func(c chan *addr.HeightOutput) {
			for msg := range c {
				aggregateOutput <- msg
			}
		}(ch)
	}
	var txChan = make(chan *model.Tx)
	go func() {
		defer func() {
			close(txChan)
			cancel()
		}()
		for {
			var txHashString string
			select {
			case addrHeightOutput, ok := <-aggregateOutput:
				if !ok {
					return
				}
				txHashString = chainhash.Hash(addrHeightOutput.TxHash).String()
			case addrHeightInput, ok := <-aggregateInput:
				if !ok {
					return
				}
				txHashString = chainhash.Hash(addrHeightInput.TxHash).String()
			}
			raw, err := dataloader.NewTxRawLoader(txRawLoaderConfig).Load(txHashString)
			if err != nil {
				jerr.Get("error getting tx raw from dataloader for address subscription resolver", err).Print()
				return
			}
			txChan <- &model.Tx{
				Hash: txHashString,
				Raw:  raw,
			}
		}
	}()
	return txChan, nil
}

// Blocks is the resolver for the blocks field.
func (r *subscriptionResolver) Blocks(ctx context.Context) (<-chan *model.Block, error) {
	ctx, cancel := context.WithCancel(ctx)
	blockHeightListener, err := chain.ListenBlockHeights(ctx)
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting block height listener for subscription", err)
	}
	var blockChan = make(chan *model.Block)
	go func() {
		defer func() {
			close(blockChan)
			cancel()
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
			block, err := chain.GetBlock(blockHeight.BlockHash[:])
			if err != nil {
				jerr.Get("error getting block for block height subscription", err).Print()
				return
			}
			blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
			if err != nil {
				jerr.Get("error getting block header from raw", err).Print()
				return
			}
			height := int(blockHeight.Height)
			blockChan <- &model.Block{
				Hash:      chainhash.Hash(blockHeight.BlockHash).String(),
				Timestamp: model.Date(blockHeader.Timestamp),
				Height:    &height,
			}
		}
	}()
	return blockChan, nil
}

// Posts is the resolver for the posts field.
func (r *subscriptionResolver) Posts(ctx context.Context, hashes []string) (<-chan *model.Post, error) {
	postChan, err := new(sub.Post).Listen(ctx, hashes)
	if err != nil {
		return nil, jerr.Get("error getting post listener for subscription", err)
	}
	return postChan, nil
}

// Profiles is the resolver for the profiles field.
func (r *subscriptionResolver) Profiles(ctx context.Context, addresses []string) (<-chan *model.Profile, error) {
	var profile = new(sub.Profile)
	profileChan, err := profile.Listen(ctx, addresses, GetPreloads(ctx))
	if err != nil {
		return nil, jerr.Get("error getting profile listener for subscription", err)
	}
	return profileChan, nil
}

// Rooms is the resolver for the rooms field.
func (r *subscriptionResolver) Rooms(ctx context.Context, names []string) (<-chan *model.Post, error) {
	var room = new(sub.Room)
	roomPostsChan, err := room.Listen(ctx, names)
	if err != nil {
		return nil, jerr.Get("error getting room listener for subscription", err)
	}
	return roomPostsChan, nil
}

// RoomFollows is the resolver for the room_follows field.
func (r *subscriptionResolver) RoomFollows(ctx context.Context, addresses []string) (<-chan *model.RoomFollow, error) {
	var roomFollow = new(sub.RoomFollow)
	roomFollowsChan, err := roomFollow.Listen(ctx, addresses)
	if err != nil {
		return nil, jerr.Get("error getting room follow listener for subscription", err)
	}
	return roomFollowsChan, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
