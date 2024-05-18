package sub

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type RoomFollow struct {
	Name   string
	Cancel context.CancelFunc
}

func (r *RoomFollow) Listen(ctx context.Context, addresses []string) (<-chan *model.RoomFollow, error) {
	var addrs [][25]byte
	for _, address := range addresses {
		addr, err := wallet.GetAddrFromString(address)
		if err != nil {
			return nil, jerr.Get("error getting addr for room follow subscription", err)
		}
		addrs = append(addrs, *addr)
	}
	ctx, r.Cancel = context.WithCancel(ctx)
	var roomFollowChan = make(chan *model.RoomFollow)
	lockRoomFollowListener, err := memo.ListenAddrRoomFollows(ctx, addrs)
	if err != nil {
		r.Cancel()
		return nil, jerr.Get("error getting memo lock room follow listener for room follow subscription", err)
	}
	go func() {
		defer func() {
			close(roomFollowChan)
			r.Cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case lockRoomFollow, ok := <-lockRoomFollowListener:
				if !ok {
					return
				}
				roomFollowChan <- &model.RoomFollow{
					Name:     lockRoomFollow.Room,
					Address:  wallet.Addr(lockRoomFollow.Addr).String(),
					Unfollow: lockRoomFollow.Unfollow,
					TxHash:   chainhash.Hash(lockRoomFollow.TxHash).String(),
				}
			}
		}
	}()
	return roomFollowChan, nil
}
