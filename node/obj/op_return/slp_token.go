package op_return

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var slpTokenHandler = &Handler{
	prefix: memo.PrefixSlp,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) < 5 {
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("invalid slp, incorrect push data (%d) op return handler", len(info.PushData)),
			}); err != nil {
				return jerr.Get("error saving process error for slp incorrect push data", err)
			}
			return nil
		}
		switch memo.SlpType(info.PushData[2]) {
		case memo.SlpTxTypeGenesis:
			if err := save.SlpGenesis(info); err != nil {
				return jerr.Get("error saving slp genesis op return handler", err)
			}
		case memo.SlpTxTypeMint:
			if err := save.SlpMint(info); err != nil {
				return jerr.Get("error saving slp mint op return handler", err)
			}
		case memo.SlpTxTypeSend:
			if err := save.SlpSend(info); err != nil {
				return jerr.Get("error saving slp send op return handler", err)
			}
		case memo.SlpTxTypeCommit:
			if err := save.SlpCommit(info); err != nil {
				return jerr.Get("error saving slp commit op return handler", err)
			}
		default:
			if err := item.LogProcessError(&item.ProcessError{
				TxHash: info.TxHash,
				Error:  fmt.Sprintf("unknown slp tx type op return handler: %s", info.PushData[2]),
			}); err != nil {
				return jerr.Get("error saving process error for slp unknown tx type", err)
			}
			return nil
		}
		return nil
	},
}
