package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoProfilePicHandler = &Handler{
	prefix: memo.PrefixSetProfilePic,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set profile pic, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error", err)
			}
			return nil
		}
		var pic = jutil.GetUtf8String(info.PushData[1])
		var lockMemoProfilePic = &item.LockMemoProfilePic{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Pic:      pic,
		}
		if err := item.Save([]item.Object{lockMemoProfilePic}); err != nil {
			return jerr.Get("error saving db lock memo profile pic object", err)
		}
		if info.Height != item.HeightMempool {
			lockMemoProfilePic.Height = item.HeightMempool
			if err := item.RemoveLockMemoProfilePic(lockMemoProfilePic); err != nil {
				return jerr.Get("error removing db lock memo profile pic", err)
			}
		}
		return nil
	},
}
