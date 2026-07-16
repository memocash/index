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
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var memoLinkRequestHandler = &Handler{
	prefix: memo.PrefixLinkRequest,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 2 && len(info.PushData) != 3 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid link request, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link request incorrect push data; %w", err)
			}
			return nil
		}
		parentAddress, err := wallet.GetAddressFromPkHashNew(info.PushData[1])
		if err != nil {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error getting address from link request parent pk hash; %s", err),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link request parent address; %w", err)
			}
			return nil
		}
		var message string
		if len(info.PushData) == 3 {
			message = jutil.GetUtf8String(info.PushData[2])
		}
		var addrLinkRequest = &dbMemo.AddrLinkRequest{
			Addr:       info.Addr,
			Seen:       info.Seen,
			TxHash:     info.TxHash,
			ParentAddr: parentAddress.GetAddr(),
			Message:    message,
		}
		if err := db.Save([]db.Object{addrLinkRequest}); err != nil {
			return fmt.Errorf("error saving db memo link request object; %w", err)
		}
		return nil
	},
}

var memoLinkAcceptHandler = &Handler{
	prefix: memo.PrefixLinkAccept,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 2 && len(info.PushData) != 3 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid link accept, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link accept incorrect push data; %w", err)
			}
			return nil
		}
		if len(info.PushData[1]) != memo.TxHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error link accept request tx hash not correct size: %d", len(info.PushData[1])),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link accept request tx hash; %w", err)
			}
			return nil
		}
		var requestTxHash [32]byte
		copy(requestTxHash[:], jutil.ByteReverse(info.PushData[1]))
		var message string
		if len(info.PushData) == 3 {
			message = jutil.GetUtf8String(info.PushData[2])
		}
		var addrLinkAccept = &dbMemo.AddrLinkAccept{
			Addr:          info.Addr,
			Seen:          info.Seen,
			TxHash:        info.TxHash,
			RequestTxHash: requestTxHash,
			Message:       message,
		}
		if err := db.Save([]db.Object{addrLinkAccept}); err != nil {
			return fmt.Errorf("error saving db memo link accept object; %w", err)
		}
		return nil
	},
}

var memoLinkRevokeHandler = &Handler{
	prefix: memo.PrefixLinkRevoke,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		if len(info.PushData) != 2 && len(info.PushData) != 3 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid link revoke, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link revoke incorrect push data; %w", err)
			}
			return nil
		}
		if len(info.PushData[1]) != memo.TxHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error link revoke accept tx hash not correct size: %d", len(info.PushData[1])),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link revoke accept tx hash; %w", err)
			}
			return nil
		}
		var acceptTxHash [32]byte
		copy(acceptTxHash[:], jutil.ByteReverse(info.PushData[1]))
		var message string
		if len(info.PushData) == 3 {
			message = jutil.GetUtf8String(info.PushData[2])
		}
		var addrLinkRevoke = &dbMemo.AddrLinkRevoke{
			Addr:         info.Addr,
			Seen:         info.Seen,
			TxHash:       info.TxHash,
			AcceptTxHash: acceptTxHash,
			Message:      message,
		}
		if err := db.Save([]db.Object{addrLinkRevoke}); err != nil {
			return fmt.Errorf("error saving db memo link revoke object; %w", err)
		}
		return nil
	},
}
