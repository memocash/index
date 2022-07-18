package op_return

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoFollowHandler = &Handler{
	prefix: memo.PrefixFollow,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			return jerr.Newf("invalid set follow, incorrect push data (%d)", len(info.PushData))
		}
		var memoFollow = &item.MemoFollow{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Follow:   info.PushData[1],
		}
		if err := item.Save([]item.Object{memoFollow}); err != nil {
			return jerr.Get("error saving db memo follow object", err)
		}
		if info.Height != item.HeightMempool {
			memoFollow.Height = item.HeightMempool
			if err := item.RemoveMemoFollow(memoFollow); err != nil {
				return jerr.Get("error removing db memo follow", err)
			}
		}
		return nil
	},
}
