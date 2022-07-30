package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoProfileHandler = &Handler{
	prefix: memo.PrefixSetProfile,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set profile, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error", err)
			}
			return nil
		}
		var profile = jutil.GetUtf8String(info.PushData[1])
		var memoProfile = &item.MemoProfile{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Profile:  profile,
		}
		if err := item.Save([]item.Object{memoProfile}); err != nil {
			return jerr.Get("error saving db memo profile object", err)
		}
		if info.Height != item.HeightMempool {
			memoProfile.Height = item.HeightMempool
			if err := item.RemoveMemoProfile(memoProfile); err != nil {
				return jerr.Get("error removing db memo profile", err)
			}
		}
		return nil
	},
}
