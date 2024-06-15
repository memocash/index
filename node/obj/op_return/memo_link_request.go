package op_return

import (
	"context"
	"fmt"
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
		if len(info.PushData) < 2 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid set link request, incorrect push data (%d)", len(info.PushData)),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link request incorrect push data; %w", err)
			}
			return nil
		}
		if len(info.PushData[1]) != memo.PkHashLength {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("error link request address incorrect length: %d", len(info.PushData[1])),
			}); err != nil {
				return fmt.Errorf("error saving process error memo link request address; %w", err)
			}
			return nil
		}
		parentAddr := *wallet.GetAddrFromPkHash(info.PushData[1])
		var message string
		if len(info.PushData) > 2 {
			message = string(info.PushData[2])
		}
		var linkRequest = &dbMemo.LinkRequest{
			TxHash:     info.TxHash,
			ChildAddr:  info.Addr,
			ParentAddr: parentAddr,
			Message:    message,
		}
		var addrLinkRequest = &dbMemo.AddrLinkRequest{
			Addr:   info.Addr,
			Seen:   info.Seen,
			TxHash: info.TxHash,
		}
		var addrLinkRequested = &dbMemo.AddrLinkRequested{
			Addr:   parentAddr,
			Seen:   info.Seen,
			TxHash: info.TxHash,
		}
		if err := db.Save([]db.Object{linkRequest, addrLinkRequest, addrLinkRequested}); err != nil {
			return fmt.Errorf("error saving db lock memo link request object; %w", err)
		}
		return nil
	},
}
