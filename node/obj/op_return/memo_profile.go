package op_return

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
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
		jlog.Log("profile handler")
		var profile = jutil.GetUtf8String(info.PushData[1])
		var setProfile = &item.MemoProfile{
			LockHash: info.LockHash,
			Height:   info.Height,
			TxHash:   info.TxHash,
			Profile:  profile,
		}
		if err := item.Save([]item.Object{setProfile}); err != nil {
			return jerr.Get("error saving db memo profile object", err)
		}
		return nil
	},
}
