package op_return

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
)

var memoProfileHandler = &Handler{
	prefix: memo.PrefixSetProfile,
	handle: func(info Info) error {
		if len(info.PushData) != 2 {
			return jerr.Newf("invalid set profile, incorrect push data (%d)", len(info.PushData))
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
