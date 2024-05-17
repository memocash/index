package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/admin/graph/sub"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	memo_db "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/db/metric"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Tx is the resolver for the tx field.
func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	metric.AddGraphQuery(metric.EndPointTx)
	tx, err := load.GetTxByString(ctx, hash)
	if err != nil {
		if errors.Is(err, load.TxMissingError) {
			return nil, fmt.Errorf("tx not found for hash: %s", hash)
		}
		return nil, fmt.Errorf("error getting tx from dataloader for tx query resolver; %w", err)
	}
	return tx, nil
}

// Txs is the resolver for the txs field.
func (r *queryResolver) Txs(ctx context.Context, hashes []string) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// Address is the resolver for the address field.
func (r *queryResolver) Address(ctx context.Context, address string) (*model.Lock, error) {
	metric.AddGraphQuery(metric.EndPointAddress)
	lock, err := load.GetLock(ctx, address)
	if err != nil {
		return nil, jerr.Getf(err, "error getting lock from loader for query address resolver: %s", address)
	}
	return lock, nil
}

// Addresses is the resolver for the addresses field.
func (r *queryResolver) Addresses(ctx context.Context, addresses []string) ([]*model.Lock, error) {
	metric.AddGraphQuery(metric.EndPointAddresses)
	if load.GetFields(ctx).HasField("balance") {
		// TODO: Reimplement if needed
		return nil, jerr.New("error balance no longer implemented")
	}
	var locks []*model.Lock
	for _, address := range addresses {
		locks = append(locks, &model.Lock{
			Address: address,
		})
	}
	if err := load.AttachToLocks(ctx, load.GetFields(ctx), locks); err != nil {
		return nil, jerr.Get("error attaching to locks for query resolver", err)
	}
	return locks, nil
}

// Block is the resolver for the block field.
func (r *queryResolver) Block(ctx context.Context, hash string) (*model.Block, error) {
	metric.AddGraphQuery(metric.EndPointBlock)
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
		Hash:      blockHeight.BlockHash,
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
		Raw:       block.Raw,
	}
	if err := load.AttachToBlocks(ctx, load.GetFields(ctx), []*model.Block{modelBlock}); err != nil {
		return nil, jerr.Get("error attaching to block for query resolver", err)
	}
	return modelBlock, nil
}

// BlockNewest is the resolver for the block_newest field.
func (r *queryResolver) BlockNewest(ctx context.Context) (*model.Block, error) {
	metric.AddGraphQuery(metric.EndPointBlockNewest)
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
		Hash:      heightBlock.BlockHash,
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
	}, nil
}

// Blocks is the resolver for the blocks field.
func (r *queryResolver) Blocks(ctx context.Context, newest *bool, start *uint32) ([]*model.Block, error) {
	metric.AddGraphQuery(metric.EndPointBlocks)
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
	var blockHashes = make([][32]byte, len(heightBlocks))
	for i := range heightBlocks {
		blockHashes[i] = heightBlocks[i].BlockHash
	}
	blocks, err := chain.GetBlocks(ctx, blockHashes)
	if err != nil {
		return nil, jerr.Get("error getting raw blocks", err)
	}
	var modelBlocks = make([]*model.Block, len(heightBlocks))
	for i := range heightBlocks {
		var height = int(heightBlocks[i].Height)
		modelBlocks[i] = &model.Block{
			Hash:   heightBlocks[i].BlockHash,
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
	}
	if err := load.AttachToBlocks(ctx, load.GetFields(ctx), modelBlocks); err != nil {
		return nil, jerr.Get("error attaching to blocks for query resolver", err)

	}
	return modelBlocks, nil
}

// Profiles is the resolver for the profiles field.
func (r *queryResolver) Profiles(ctx context.Context, addresses []string) ([]*model.Profile, error) {
	metric.AddGraphQuery(metric.EndPointProfiles)
	var profiles []*model.Profile
	for _, addressString := range addresses {
		profile, err := load.GetProfile(ctx, addressString)
		if err != nil {
			return nil, jerr.Get("error getting profile from dataloader for profile query resolver", err)
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context, txHashes []string) ([]*model.Post, error) {
	metric.AddGraphQuery(metric.EndPointPosts)
	posts, errs := load.Post.LoadAll(txHashes)
	for i, err := range errs {
		if err != nil {
			return nil, jerr.Getf(err, "error getting post from post dataloader for query resolver: %s", txHashes[i])
		}
	}
	return posts, nil
}

// PostsNewest is the resolver for the posts_newest field.
func (r *queryResolver) PostsNewest(ctx context.Context, start *model.Date, tx *string) ([]*model.Post, error) {
	metric.AddGraphQuery(metric.EndPointPostsNewest)
	var txHash chainhash.Hash
	if tx != nil {
		txHashPointer, _ := chainhash.NewHashFromStr(*tx)
		txHash = *txHashPointer
	}
	var startTime time.Time
	if start != nil {
		startTime = time.Time(*start)
	}
	if tx != nil && start == nil {
		txSeens, err := chain.GetTxSeens(ctx, [][32]byte{txHash})
		if err != nil {
			return nil, jerr.Get("error getting tx seen param for newest graphql query", err)
		}
		if len(txSeens) > 0 {
			startTime = txSeens[0].Timestamp
		}
	}
	seenPosts, err := memo_db.GetSeenPosts(ctx, startTime, txHash)
	if err != nil {
		return nil, jerr.Get("error getting seen posts for newest graphql query", err)
	}
	var txHashes = make([]string, len(seenPosts))
	for i := range seenPosts {
		txHashes[i] = chainhash.Hash(seenPosts[i].PostTxHash).String()
	}
	posts, errs := load.Post.LoadAll(txHashes)
	for i, err := range errs {
		if err != nil {
			return nil, jerr.Getf(err, "error getting post from post dataloader for query resolver: %s", txHashes[i])
		}
	}
	var returnPosts []*model.Post
	for i := range txHashes {
		for _, post := range posts {
			if post.TxHash == txHashes[i] {
				returnPosts = append(returnPosts, post)
				break
			}
		}
	}
	return returnPosts, nil
}

// Room is the resolver for the room field.
func (r *queryResolver) Room(ctx context.Context, name string) (*model.Room, error) {
	metric.AddGraphQuery(metric.EndPointRoom)
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
	addrSeenTxsListener, err := addr.ListenAddrSeenTxs(ctx, addrs)
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting addr seen txs listener for address subscription", err)
	}
	fields := load.GetFields(ctx)
	var txChan = make(chan *model.Tx)
	go func() {
		defer func() {
			close(txChan)
			cancel()
		}()
		for {
			addrSeenTx, ok := <-addrSeenTxsListener
			if !ok {
				return
			}
			var tx = &model.Tx{
				Hash: addrSeenTx.TxHash,
				Seen: model.Date(addrSeenTx.Seen),
			}
			if err := load.AttachToTxs(ctx, fields, []*model.Tx{tx}); err != nil {
				log.Printf("error attaching to txs for address subscription; %v", err)
				return
			}
			txChan <- tx
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
			block, err := chain.GetBlock(blockHeight.BlockHash)
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
				Hash:      blockHeight.BlockHash,
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
	profileChan, err := profile.Listen(ctx, addresses, load.GetFields(ctx))
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
