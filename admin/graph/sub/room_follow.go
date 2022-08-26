package sub

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type RoomFollow struct {
	Name   string
	Cancel context.CancelFunc
}

func (r *RoomFollow) Listen(ctx context.Context, addresses []string) (<-chan *model.RoomFollow, error) {
	var lockHashes [][]byte
	for _, address := range addresses {
		lockScript, err := get.LockScriptFromAddress(wallet.GetAddressFromString(address))
		if err != nil {
			return nil, jerr.Get("error getting lock script for room follow subscription", err)
		}
		lockHashes = append(lockHashes, script.GetLockHash(lockScript))
	}
	ctx, r.Cancel = context.WithCancel(ctx)
	var roomFollowChan = make(chan *model.RoomFollow)
	lockRoomFollowListener, err := memo.ListenLockHeightRoomFollows(ctx, lockHashes)
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
					LockHash: hex.EncodeToString(lockRoomFollow.LockHash),
					Unfollow: lockRoomFollow.Unfollow,
					TxHash:   hs.GetTxString(lockRoomFollow.TxHash),
				}
			}
		}
	}()
	return roomFollowChan, nil
}
