package op_return

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	dbMemo "github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var memoSendHandler = &Handler{
	prefix: memo.PrefixSendMoney,
	handle: func(ctx context.Context, info parse.OpReturn) error {
		message, err := getMemoSendMessage(info.PushData)
		if err != nil {
			if logErr := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  err.Error(),
			}); logErr != nil {
				return fmt.Errorf("error saving process error for memo send; %w", logErr)
			}
			return nil
		}
		// Persist the primary post before optional recipient enrichment. A valid
		// Send remains a post even when its payment output cannot be resolved.
		if err := save.MemoPost(ctx, info, message); err != nil {
			return fmt.Errorf("error saving memo post for memo send handler; %w", err)
		}
		recipient, err := getMemoSendRecipient(info.PushData[1], info.Outputs)
		if err != nil {
			if logErr := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  err.Error(),
			}); logErr != nil {
				return fmt.Errorf("error saving process error for memo send recipient; %w", logErr)
			}
			return nil
		}
		if err := db.Save([]db.Object{&dbMemo.Send{
			TxHash:    info.TxHash,
			Recipient: [25]byte(recipient),
		}}); err != nil {
			return fmt.Errorf("error saving memo send recipient; %w", err)
		}
		return nil
	},
}

func getMemoSendMessage(pushData [][]byte) (string, error) {
	if len(pushData) != 3 {
		return "", fmt.Errorf("invalid send, incorrect push data (%d)", len(pushData))
	}
	if len(pushData[1]) != memo.PkHashLength &&
		len(pushData[1]) != memo.ScriptHashLength {
		return "", fmt.Errorf("invalid send, incorrect recipient hash length (%d)", len(pushData[1]))
	}
	if len(pushData[2]) == 0 {
		return "", fmt.Errorf("invalid send, empty message")
	}
	return jutil.GetUtf8String(pushData[2]), nil
}

func getMemoSendRecipient(hash []byte, outputs []*wire.TxOut) (wallet.Addr, error) {
	for _, output := range outputs {
		addr, err := wallet.GetAddrFromLockScript(output.PkScript)
		if err == nil && addr != nil && bytes.Equal(addr[1:21], hash) {
			return *addr, nil
		}
	}
	return wallet.Addr{}, fmt.Errorf("invalid send, recipient output not found")
}
