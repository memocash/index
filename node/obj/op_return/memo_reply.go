package op_return

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
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
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 3 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid reply, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error for memo reply incorrect push data; %w", err)
			}
			return nil
		}
		parentTxHash, err := chainhash.NewHash(info.PushData[1])
		if err != nil {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid parent tx hash for reply (%x); %s", info.PushData[1], err),
			}); err != nil {
				return fmt.Errorf("error saving process error for memo reply invalid parent tx hash; %w", err)
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
			return fmt.Errorf("error saving memo post parent and child for memo reply handler; %w", err)
		}
		var post = jutil.GetUtf8String(info.PushData[2])
		if err := save.MemoPost(ctx, info, post); err != nil {
			return fmt.Errorf("error saving memo post for memo reply handler; %w", err)
		}
		return nil
	},
}
