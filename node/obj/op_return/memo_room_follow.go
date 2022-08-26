package op_return

import (
	"bytes"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoRoomFollowHandler = &Handler{
	prefix: memo.PrefixTopicFollow,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid chat room follow, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error for memo chat room follow incorrect push data", err)
			}
			return nil
		}
		unfollow := bytes.Equal(info.PushData[0], memo.PrefixTopicUnfollow)
		var room = jutil.GetUtf8String(info.PushData[1])
		roomHash := dbMemo.GetRoomHash(room)
		var lockRoomFollow = &dbMemo.LockHeightRoomFollow{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Room:     room,
			Unfollow: unfollow,
		}
		var roomFollow = &dbMemo.RoomHeightFollow{
			RoomHash: roomHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			LockHash: info.LockHash,
			Unfollow: unfollow,
		}
		if err := db.Save([]db.Object{lockRoomFollow, roomFollow}); err != nil {
			return jerr.Get("error saving db memo room height follow objects", err)
		}
		if info.Height != item.HeightMempool {
			lockRoomFollow.Height = item.HeightMempool
			roomFollow.Height = item.HeightMempool
			if err := db.Remove([]db.Object{lockRoomFollow, roomFollow}); err != nil {
				return jerr.Get("error removing db memo room height follow objects", err)
			}
		}
		return nil
	},
}

var memoRoomUnfollowHandler = &Handler{
	prefix: memo.PrefixTopicUnfollow,
	handle: memoRoomFollowHandler.handle,
}
