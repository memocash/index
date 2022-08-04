package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoLikeHandler = &Handler{
	prefix: memo.PrefixLike,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set like, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error memo like incorrect push data", err)
			}
			return nil
		}
		likeTxHash := info.PushData[1]
		if len(likeTxHash) != memo.TxHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error like tx hash not correct size: %d", len(likeTxHash)),
			}); err != nil {
				return jerr.Get("error saving process error memo like address", err)
			}
			return nil
		}
		var memoLike = &item.MemoLike{
			LockHash:   info.LockHash,
			Height:     info.Height,
			TxHash:     info.TxHash,
			LikeTxHash: likeTxHash,
		}
		if err := item.Save([]item.Object{memoLike}); err != nil {
			return jerr.Get("error saving db memo like object", err)
		}
		if info.Height != item.HeightMempool {
			memoLike.Height = item.HeightMempool
			if err := item.RemoveMemoLike(memoLike); err != nil {
				return jerr.Get("error removing db memo like", err)
			}
		}
		return nil
	},
}
