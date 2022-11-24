package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoProfileHandler = &Handler{
	prefix: memo.PrefixSetProfile,
	handle: func(info parse.OpReturn) error {
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
		var addrMemoProfile = &dbMemo.AddrHeightProfile{
			Addr:    info.Addr,
			Height:  info.Height,
			TxHash:  info.TxHash,
			Profile: profile,
		}
		if err := db.Save([]db.Object{addrMemoProfile}); err != nil {
			return jerr.Get("error saving db addr memo profile object", err)
		}
		if info.Height != item.HeightMempool {
			addrMemoProfile.Height = item.HeightMempool
			if err := db.Remove([]db.Object{addrMemoProfile}); err != nil {
				return jerr.Get("error removing db addr memo profile", err)
			}
		}
		return nil
	},
}
