package op_return

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoProfilePicHandler = &Handler{
	prefix: memo.PrefixSetProfilePic,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			return jerr.Newf("invalid set profile pic, incorrect push data (%d)", len(info.PushData))
		}
		var pic = jutil.GetUtf8String(info.PushData[1])
		var memoProfilePic = &item.MemoProfilePic{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Pic:      pic,
		}
		if err := item.Save([]item.Object{memoProfilePic}); err != nil {
			return jerr.Get("error saving db memo profile pic object", err)
		}
		if info.Height != item.HeightMempool {
			memoProfilePic.Height = item.HeightMempool
			if err := item.RemoveMemoProfilePic(memoProfilePic); err != nil {
				return jerr.Get("error removing db memo profile pic", err)
			}
		}
		return nil
	},
}
