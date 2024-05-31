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
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	memo_db "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/db/metric"
	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/graph/sub"
	"github.com/memocash/index/ref/bitcoin/memo"
)

// Tx is the resolver for the tx field.
func (r *queryResolver) Tx(ctx context.Context, hash model.Hash) (*model.Tx, error) {
	metric.AddGraphQuery(metric.EndPointTx)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var tx = &model.Tx{Hash: hash}
	if err := attach.ToTxs(ctxWithTimeout, attach.GetFields(ctx), []*model.Tx{tx}); err != nil {
		if errors.Is(err, attach.TxMissingError) {
			return nil, fmt.Errorf("tx not found: %s", hash)
		}
		return nil, InternalError{fmt.Errorf("error attaching to tx for query resolver; %w", err)}
	}
	return tx, nil
}

// Txs is the resolver for the txs field.
func (r *queryResolver) Txs(ctx context.Context, hashes []model.Hash) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

// Address is the resolver for the address field.
func (r *queryResolver) Address(ctx context.Context, address model.Address) (*model.Lock, error) {
	metric.AddGraphQuery(metric.EndPointAddress)
	var lock = &model.Lock{Address: address}
	if err := attach.ToLocks(ctx, attach.GetFields(ctx), []*model.Lock{lock}); err != nil {
		return nil, fmt.Errorf("error attaching details to lock; %w", err)
	}
	return lock, nil
}

// Addresses is the resolver for the addresses field.
func (r *queryResolver) Addresses(ctx context.Context, addresses []model.Address) ([]*model.Lock, error) {
	metric.AddGraphQuery(metric.EndPointAddresses)
	if attach.GetFields(ctx).HasField("balance") {
		// TODO: Reimplement if needed
		return nil, InternalError{fmt.Errorf("error balance no longer implemented")}
	}
	var locks []*model.Lock
	for _, address := range addresses {
		locks = append(locks, &model.Lock{
			Address: address,
		})
	}
	if err := attach.ToLocks(ctx, attach.GetFields(ctx), locks); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to locks for query resolver; %w", err)}
	}
	return locks, nil
}

// Block is the resolver for the block field.
func (r *queryResolver) Block(ctx context.Context, hash model.Hash) (*model.Block, error) {
	metric.AddGraphQuery(metric.EndPointBlock)
	blockHeight, err := chain.GetBlockHeight(hash)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting block height for query resolver; %w", err)}
	}
	block, err := chain.GetBlock(hash)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting raw block; %w", err)}
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting block header from raw; %w", err)}
	}
	height := int(blockHeight.Height)
	var modelBlock = &model.Block{
		Hash:      blockHeight.BlockHash,
		Timestamp: model.Date(blockHeader.Timestamp),
		Height:    &height,
		Raw:       block.Raw,
	}
	if err := attach.ToBlocks(ctx, attach.GetFields(ctx), []*model.Block{modelBlock}); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to block for query resolver; %w", err)}
	}
	return modelBlock, nil
}

// BlockNewest is the resolver for the block_newest field.
func (r *queryResolver) BlockNewest(ctx context.Context) (*model.Block, error) {
	metric.AddGraphQuery(metric.EndPointBlockNewest)
	heightBlock, err := chain.GetRecentHeightBlock()
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting recent height block for query; %w", err)}
	}
	if heightBlock == nil {
		return nil, nil
	}
	block, err := chain.GetBlock(heightBlock.BlockHash)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting raw block; %w", err)}
	}
	blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting block header from raw; %w", err)}
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
		return nil, InternalError{fmt.Errorf("error getting height blocks for query; %w", err)}
	}
	var blockHashes = make([][32]byte, len(heightBlocks))
	for i := range heightBlocks {
		blockHashes[i] = heightBlocks[i].BlockHash
	}
	blocks, err := chain.GetBlocks(ctx, blockHashes)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting raw blocks; %w", err)}
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
					return nil, InternalError{fmt.Errorf("error getting block header from raw; %w", err)}
				}
				modelBlocks[i].Timestamp = model.Date(blockHeader.Timestamp)
			}
		}
	}
	if err := attach.ToBlocks(ctx, attach.GetFields(ctx), modelBlocks); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to blocks for query resolver; %w", err)}

	}
	return modelBlocks, nil
}

// Profiles is the resolver for the profiles field.
func (r *queryResolver) Profiles(ctx context.Context, addresses []model.Address) ([]*model.Profile, error) {
	metric.AddGraphQuery(metric.EndPointProfiles)
	var profiles []*model.Profile
	for _, address := range addresses {
		profiles = append(profiles, &model.Profile{Address: address})
	}
	if err := attach.ToMemoProfiles(ctx, attach.GetFields(ctx), profiles); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to profiles for query resolver; %w", err)}
	}
	return profiles, nil
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context, txHashes []model.Hash) ([]*model.Post, error) {
	metric.AddGraphQuery(metric.EndPointPosts)
	var posts []*model.Post
	for _, txHash := range txHashes {
		posts = append(posts, &model.Post{TxHash: txHash})
	}
	if err := attach.ToMemoPosts(ctx, attach.GetFields(ctx), posts); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to posts for query resolver posts; %w", err)}
	}
	return posts, nil
}

// PostsNewest is the resolver for the posts_newest field.
func (r *queryResolver) PostsNewest(ctx context.Context, start *model.Date, tx *model.Hash, limit *uint32) ([]*model.Post, error) {
	metric.AddGraphQuery(metric.EndPointPostsNewest)
	var txHash chainhash.Hash
	if tx != nil {
		txHash = chainhash.Hash(*tx)
	}
	var startTime time.Time
	if start != nil {
		startTime = time.Time(*start)
	}
	if tx != nil && start == nil {
		txSeens, err := chain.GetTxSeens(ctx, [][32]byte{txHash})
		if err != nil {
			return nil, InternalError{fmt.Errorf("error getting tx seen param for newest graphql query; %w", err)}
		}
		if len(txSeens) > 0 {
			startTime = txSeens[0].Timestamp
		}
	}
	var limitInt uint32
	if limit != nil {
		limitInt = *limit
	}
	seenPosts, err := memo_db.GetSeenPosts(ctx, startTime, txHash, limitInt)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting seen posts for newest graphql query; %w", err)}
	}
	var posts []*model.Post
	for _, seenPost := range seenPosts {
		posts = append(posts, &model.Post{TxHash: seenPost.PostTxHash})
	}
	if err := attach.ToMemoPosts(ctx, attach.GetFields(ctx), posts); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to posts for query resolver posts newest; %w", err)}
	}
	return posts, nil
}

// Room is the resolver for the room field.
func (r *queryResolver) Room(ctx context.Context, name string) (*model.Room, error) {
	metric.AddGraphQuery(metric.EndPointRoom)
	var room = &model.Room{Name: name}
	if err := attach.ToMemoRooms(ctx, attach.GetFields(ctx), []*model.Room{room}); err != nil {
		return nil, InternalError{fmt.Errorf("error attaching to rooms for room query resolver; %w", err)}
	}
	return room, nil
}

// Address is the resolver for the address field.
func (r *subscriptionResolver) Address(ctx context.Context, address model.Address) (<-chan *model.Tx, error) {
	txChan, err := r.Addresses(ctx, []model.Address{address})
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting address for address subscription; %w", err)}
	}
	return txChan, nil
}

// Addresses is the resolver for the address field.
func (r *subscriptionResolver) Addresses(ctx context.Context, addresses []model.Address) (<-chan *model.Tx, error) {
	ctx, cancel := context.WithCancel(ctx)
	addrSeenTxsListener, err := addr.ListenAddrSeenTxs(ctx, model.AddressesToArrays(addresses))
	if err != nil {
		cancel()
		return nil, InternalError{fmt.Errorf("error getting addr seen txs listener for address subscription; %w", err)}
	}
	fields := attach.GetFields(ctx)
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
			if err := attach.ToTxs(ctx, fields, []*model.Tx{tx}); err != nil {
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
		return nil, InternalError{fmt.Errorf("error getting block height listener for subscription; %w", err)}
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
				log.Println(fmt.Errorf("error getting block for block height subscription; %w", err))
				return
			}
			blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
			if err != nil {
				log.Println(fmt.Errorf("error getting block header from raw; %w", err))
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
func (r *subscriptionResolver) Posts(ctx context.Context, hashes []model.Hash) (<-chan *model.Post, error) {
	postChan, err := new(sub.Post).Listen(ctx, model.HashesToArrays(hashes))
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting post listener for subscription; %w", err)}
	}
	return postChan, nil
}

// Profiles is the resolver for the profiles field.
func (r *subscriptionResolver) Profiles(ctx context.Context, addresses []model.Address) (<-chan *model.Profile, error) {
	var profile = new(sub.Profile)
	profileChan, err := profile.Listen(ctx, model.AddressesToArrays(addresses), attach.GetFields(ctx))
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting profile listener for subscription; %w", err)}
	}
	return profileChan, nil
}

// Rooms is the resolver for the rooms field.
func (r *subscriptionResolver) Rooms(ctx context.Context, names []string) (<-chan *model.Post, error) {
	var room = new(sub.Room)
	roomPostsChan, err := room.Listen(ctx, names)
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting room listener for subscription; %w", err)}
	}
	return roomPostsChan, nil
}

// RoomFollows is the resolver for the room_follows field.
func (r *subscriptionResolver) RoomFollows(ctx context.Context, addresses []model.Address) (<-chan *model.RoomFollow, error) {
	var roomFollow = new(sub.RoomFollow)
	roomFollowsChan, err := roomFollow.Listen(ctx, model.AddressesToArrays(addresses))
	if err != nil {
		return nil, InternalError{fmt.Errorf("error getting room follow listener for subscription; %w", err)}
	}
	return roomFollowsChan, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
