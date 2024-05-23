package sub

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
)

type RoomFollow struct {
	Name   string
	Cancel context.CancelFunc
}

func (r *RoomFollow) Listen(ctx context.Context, addresses [][25]byte) (<-chan *model.RoomFollow, error) {
	ctx, r.Cancel = context.WithCancel(ctx)
	var roomFollowChan = make(chan *model.RoomFollow)
	lockRoomFollowListener, err := memo.ListenAddrRoomFollows(ctx, addresses)
	if err != nil {
		r.Cancel()
		return nil, fmt.Errorf("error getting memo lock room follow listener for room follow subscription; %w", err)
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
					Address:  lockRoomFollow.Addr,
					Unfollow: lockRoomFollow.Unfollow,
					TxHash:   lockRoomFollow.TxHash,
				}
			}
		}
	}()
	return roomFollowChan, nil
}
