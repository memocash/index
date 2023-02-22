package resolver

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/admin/graph/sub"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

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
	addrSeenTxsListeners, err := addr.ListenAddrSeenTxsMultiple(ctx, addrs)
	if err != nil {
		cancel()
		return nil, jerr.Get("error getting addr seen txs listener for address subscription", err)
	}
	if len(addrSeenTxsListeners) == 0 {
		cancel()
		return nil, jerr.New("error no addr seen txs listeners for address subscription")
	}
	var txChan = make(chan *model.Tx)
	go func() {
		defer func() {
			close(txChan)
			cancel()
		}()
		var aggregator = make(chan *addr.SeenTx)
		for _, ch := range addrSeenTxsListeners {
			go func(c chan *addr.SeenTx) {
				for msg := range c {
					aggregator <- msg
				}
			}(ch)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-aggregator:
				if msg == nil {
					return
				}
				txChan <- &model.Tx{
					Hash: chainhash.Hash(msg.TxHash).String(),
					Seen: model.Date(msg.Seen),
				}
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

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type subscriptionResolver struct{ *Resolver }
