package op_return

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoNameHandler = &Handler{
	prefix: memo.PrefixSetName,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set name, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error; %w", err)
			}
			return nil
		}
		var name = jutil.GetUtf8String(info.PushData[1])
		var addrMemoName = &dbMemo.AddrName{
			Addr:   info.Addr,
			Seen:   info.Seen,
			TxHash: info.TxHash,
			Name:   name,
		}
		if err := db.Save([]db.Object{addrMemoName}); err != nil {
			return fmt.Errorf("error saving db memo name object; %w", err)
		}
		return nil
	},
}
