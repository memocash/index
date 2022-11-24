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

var memoNameHandler = &Handler{
	prefix: memo.PrefixSetName,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set name, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error", err)
			}
			return nil
		}
		var name = jutil.GetUtf8String(info.PushData[1])
		var addrMemoName = &dbMemo.AddrHeightName{
			Addr:   info.Addr,
			Height: info.Height,
			TxHash: info.TxHash,
			Name:   name,
		}
		if err := db.Save([]db.Object{addrMemoName}); err != nil {
			return jerr.Get("error saving db memo name object", err)
		}
		if info.Height != item.HeightMempool {
			addrMemoName.Height = item.HeightMempool
			if err := db.Remove([]db.Object{addrMemoName}); err != nil {
				return jerr.Get("error removing db memo name", err)
			}
		}
		return nil
	},
}
