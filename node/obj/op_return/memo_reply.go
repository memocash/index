package op_return

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoReplyHandler = &Handler{
	prefix: memo.PrefixReply,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) != 3 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid reply, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error for memo reply incorrect push data", err)
			}
			return nil
		}
		parentTxHash, err := chainhash.NewHash(info.PushData[1])
		if err != nil {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid parent tx hash for reply (%x)", info.PushData[1]),
			}); err != nil {
				return jerr.Get("error saving process error for memo reply invalid parent tx hash", err)
			}
			return nil
		}
		var memoPostParent = &dbMemo.PostParent{
			PostTxHash:   info.TxHash,
			ParentTxHash: *parentTxHash,
		}
		var memoPostChild = &dbMemo.PostChild{
			PostTxHash:  *parentTxHash,
			ChildTxHash: info.TxHash,
		}
		if err := db.Save([]db.Object{memoPostParent, memoPostChild}); err != nil {
			return jerr.Get("error saving memo post parent and child for memo reply handler", err)
		}
		var post = jutil.GetUtf8String(info.PushData[2])
		if err := save.MemoPost(info, post); err != nil {
			return jerr.Get("error saving memo post for memo reply handler", err)
		}
		return nil
	},
}
