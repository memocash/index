package sub

import (
	"context"
	"fmt"
	"log"

	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// RoomFollow streams room follows and unfollows by a set of addresses. Memo
// activity is keyed by the 20-byte form, so a P2SH32 address is followed
// through its P2SH20 alias and reported under the address the client asked
// for (under each, if both forms were asked for); one with no alias yet has
// its link watched and is followed once its first spend writes one.
type RoomFollow struct {
	Name      string
	Cancel    context.CancelFunc
	Fields    attach.Fields
	events    chan *memo.AddrRoomFollow
	memoAddrs *attach.MemoAddresses
}

func (r *RoomFollow) Listen(ctx context.Context, addresses []model.Address, fields attach.Fields) (<-chan *model.RoomFollow, error) {
	ctx, r.Cancel = context.WithCancel(ctx)
	r.Fields = fields
	r.events = make(chan *memo.AddrRoomFollow)
	// Watch links before resolving, so a first spend between the two is
	// still seen (by the watch rather than the lookup).
	watch, err := newLinkWatch(ctx, p2sh32Addrs(model.AddressesToLockAddrs(addresses)))
	if err != nil {
		r.Cancel()
		return nil, fmt.Errorf("error watching p2sh links for room follow subscription; %w", err)
	}
	memoAddrs, err := attach.ResolveMemoAddresses(ctx, addresses)
	if err != nil {
		r.Cancel()
		return nil, fmt.Errorf("error resolving memo addresses for room follow subscription; %w", err)
	}
	r.memoAddrs = memoAddrs
	for _, addr := range memoAddrs.Keys {
		watch.Follow(wallet.Addr(addr).LockAddr())
	}
	if len(memoAddrs.Keys) > 0 {
		if err := r.listen(ctx, memoAddrs.Keys); err != nil {
			r.Cancel()
			return nil, err
		}
	}
	if watch.Aliases != nil {
		go func() {
			// The watch ends only with the subscription or with its link
			// stream; either way the subscription is over.
			defer r.Cancel()
			for link := range watch.Aliases {
				r.memoAddrs.AddAlias(link.Other.Addr(), link.Addr)
				if !watch.Follow(link.Other) {
					// The alias was asked for directly and is already
					// followed; later rows fan out to the new form too.
					continue
				}
				if err := r.listenAlias(ctx, link.Other.Addr()); err != nil {
					log.Printf("error following new p2sh alias for room follow subscription; %v", err)
					return
				}
			}
		}()
	}
	var roomFollowChan = make(chan *model.RoomFollow)
	go func() {
		defer func() {
			close(roomFollowChan)
			r.Cancel()
		}()
		// A row can arrive both from a listener and from the catch-up read
		// of a late alias; deliver it once.
		var delivered = make(map[roomFollowKey]struct{})
		for {
			select {
			case <-ctx.Done():
				return
			case lockRoomFollow := <-r.events:
				key := roomFollowKey{Addr: lockRoomFollow.Addr, TxHash: lockRoomFollow.TxHash, Room: lockRoomFollow.Room}
				if _, ok := delivered[key]; ok {
					continue
				}
				delivered[key] = struct{}{}
				var roomFollow = &model.RoomFollow{
					Name:     lockRoomFollow.Room,
					Address:  model.NewAddress(lockRoomFollow.Addr),
					Unfollow: lockRoomFollow.Unfollow,
					TxHash:   lockRoomFollow.TxHash,
				}
				// Attached under the memo key, reported under each address
				// the client asked for it by.
				events, err := attach.ToMemoRoomFollowForms(ctx, r.Fields, roomFollow, r.memoAddrs.Addresses(lockRoomFollow.Addr))
				if err != nil {
					log.Printf("error attaching to memo room follows for room follow subscription; %v", err)
					return
				}
				for _, event := range events {
					select {
					case <-ctx.Done():
						return
					case roomFollowChan <- event:
					}
				}
			}
		}
	}()
	return roomFollowChan, nil
}

type roomFollowKey struct {
	Addr   [25]byte
	TxHash [32]byte
	Room   string
}

// listen forwards new room follow rows for addrs to the event loop. The
// listener ends only with the subscription or with a failed shard stream,
// so its end ends the subscription: the client sees the stream close and
// resubscribes, rather than keeping a stream that no longer covers addrs.
func (r *RoomFollow) listen(ctx context.Context, addrs [][25]byte) error {
	listener, err := memo.ListenAddrRoomFollows(ctx, addrs)
	if err != nil {
		return fmt.Errorf("error getting memo lock room follow listener for room follow subscription; %w", err)
	}
	go func() {
		defer r.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case lockRoomFollow, ok := <-listener:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case r.events <- lockRoomFollow:
				}
			}
		}
	}()
	return nil
}

// listenAlias follows an alias that appeared after the start, then reads
// the rows it already has: the listener is registered before the read, so a
// follow written at any point after the link (the first spend's own actions
// are saved right after its link) comes from one or the other, and the
// event loop drops the duplicate when it comes from both.
func (r *RoomFollow) listenAlias(ctx context.Context, alias [25]byte) error {
	if err := r.listen(ctx, [][25]byte{alias}); err != nil {
		return err
	}
	lockRoomFollows, err := memo.GetAddrRoomFollows(ctx, [][25]byte{alias}, 0)
	if err != nil {
		return fmt.Errorf("error getting room follows for new p2sh alias; %w", err)
	}
	for _, lockRoomFollow := range lockRoomFollows {
		select {
		case <-ctx.Done():
			return nil
		case r.events <- lockRoomFollow:
		}
	}
	return nil
}
