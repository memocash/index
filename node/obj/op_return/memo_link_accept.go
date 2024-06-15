package op_return

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var memoLinkAcceptHandler = &Handler{
	prefix: memo.PrefixLinkAccept,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) < 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set link accept, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link accept incorrect push data; %w", err)
			}
			return nil
		}
		if len(info.PushData[1]) != memo.TxHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error link accept request tx hash incorrect length: %d", len(info.PushData[1])),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link accept request tx hash; %w", err)
			}
			return nil
		}
		var acceptTxHash chainhash.Hash
		copy(acceptTxHash[:], jutil.ByteReverse(info.PushData[1]))
		var message string
		if len(info.PushData) > 2 {
			message = string(info.PushData[2])
		}
		var linkAccept = &dbMemo.LinkAccept{
			TxHash:        info.TxHash,
			Addr:          info.Addr,
			RequestTxHash: acceptTxHash,
			Message:       message,
		}
		var linkAccepted = &dbMemo.LinkAccepted{
			TxHash:        info.TxHash,
			RequestTxHash: acceptTxHash,
		}
		if err := db.Save([]db.Object{linkAccept, linkAccepted}); err != nil {
			return fmt.Errorf("error saving db lock memo link accept object; %w", err)
		}
		return nil
	},
}
