package op_return

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoPostHandler = &Handler{
	prefix: memo.PrefixPost,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid post, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error for memo post incorrect push data; %w", err)
			}
			return nil
		}
		var post = jutil.GetUtf8String(info.PushData[1])
		if err := save.MemoPost(ctx, info, post); err != nil {
			return fmt.Errorf("error saving memo post for memo post handler; %w", err)
		}
		return nil
	},
}
