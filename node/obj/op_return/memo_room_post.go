package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoRoomPostHandler = &Handler{
	prefix: memo.PrefixTopicMessage,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) != 3 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid chat room post, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error for memo chat room post incorrect push data", err)
			}
			return nil
		}
		var room = jutil.GetUtf8String(info.PushData[1])
		var post = jutil.GetUtf8String(info.PushData[2])
		if err := save.MemoPost(info, post); err != nil {
			return jerr.Get("error saving memo post for memo chat room post handler", err)
		}
		var memoRoomHeightPost = &item.MemoRoomHeightPost{
			RoomHash: item.GetMemoRoomHash(room),
			Height:   info.Height,
			TxHash:   info.TxHash,
		}
		var memoPostRoom = &item.MemoPostRoom{
			TxHash: info.TxHash,
			Room:   room,
		}
		if err := item.Save([]item.Object{memoRoomHeightPost, memoPostRoom}); err != nil {
			return jerr.Get("error saving db memo room post objects", err)
		}
		if info.Height != item.HeightMempool {
			memoRoomHeightPost.Height = item.HeightMempool
			if err := item.Remove([]item.Object{memoRoomHeightPost}); err != nil {
				return jerr.Get("error removing db memo room height post", err)
			}
		}
		return nil
	},
}
