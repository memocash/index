package op_return

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoRoomFollowHandler = &Handler{
	prefix: memo.PrefixTopicFollow,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid chat room follow, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error for memo chat room follow incorrect push data; %w", err)
			}
			return nil
		}
		unfollow := bytes.Equal(info.PushData[0], memo.PrefixTopicUnfollow)
		var room = jutil.GetUtf8String(info.PushData[1])
		roomHash := dbMemo.GetRoomHash(room)
		var lockRoomFollow = &dbMemo.AddrRoomFollow{
			Addr:     info.Addr,
			Seen:     info.Seen,
			TxHash:   info.TxHash,
			Room:     room,
			Unfollow: unfollow,
		}
		var roomFollow = &dbMemo.RoomFollow{
			RoomHash: roomHash,
			Seen:     info.Seen,
			TxHash:   info.TxHash,
			Addr:     info.Addr,
			Unfollow: unfollow,
		}
		if err := db.Save([]db.Object{lockRoomFollow, roomFollow}); err != nil {
			return fmt.Errorf("error saving db memo room height follow objects; %w", err)
		}
		return nil
	},
}

var memoRoomUnfollowHandler = &Handler{
	prefix: memo.PrefixTopicUnfollow,
	handle: memoRoomFollowHandler.handle,
}
