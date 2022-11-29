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

var memoProfilePicHandler = &Handler{
	prefix: memo.PrefixSetProfilePic,
	handle: func(info parse.OpReturn, initialSync bool) error {
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
		var addrMemoProfilePic = &dbMemo.AddrHeightProfilePic{
			Addr:   info.Addr,
			Height: info.Height,
			TxHash: info.TxHash,
			Pic:    pic,
		}
		if err := db.Save([]db.Object{addrMemoProfilePic}); err != nil {
			return jerr.Get("error saving db addr memo profile pic object", err)
		}
		if !initialSync && info.Height != item.HeightMempool {
			addrMemoProfilePic.Height = item.HeightMempool
			if err := db.Remove([]db.Object{addrMemoProfilePic}); err != nil {
				return jerr.Get("error removing db addr memo profile pic", err)
			}
		}
		return nil
	},
}
