package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoPostHandler = &Handler{
	prefix: memo.PrefixPost,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			if err := item.Save([]item.Object{&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid post, incorrect push data (%d)", len(info.PushData)),
			}}); err != nil {
				return jerr.Get("error saving process error", err)
			}
			return nil
		}
		var post = jutil.GetUtf8String(info.PushData[1])
		var memoPost = &item.MemoPost{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Post:     post,
		}
		if err := item.Save([]item.Object{memoPost}); err != nil {
			return jerr.Get("error saving db memo post object", err)
		}
		if info.Height != item.HeightMempool {
			memoPost.Height = item.HeightMempool
			if err := item.RemoveMemoPost(memoPost); err != nil {
				return jerr.Get("error removing db memo post", err)
			}
		}
		return nil
	},
}
