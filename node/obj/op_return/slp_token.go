package op_return

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/node/obj/op_return/save"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
)

var slpTokenHandler = &Handler{
	prefix: memo.PrefixPost,
	handle: func(info parse.OpReturn) error {
		if len(info.PushData) < 5 {
			return jerr.Newf("invalid slp, incorrect push data (%d) op return handler", len(info.PushData))
		}
		switch memo.SlpType(info.PushData[2]) {
		case memo.SlpTxTypeGenesis:
			err := save.SlpGenesis(info)
			if err != nil {
				return jerr.Get("error saving slp genesis op return handler", err)
			}
		case memo.SlpTxTypeMint:
			err := save.SlpMint(info)
			if err != nil {
				return jerr.Get("error saving slp mint op return handler", err)
			}
		case memo.SlpTxTypeSend:
			err := save.SlpSend(info)
			if err != nil {
				return jerr.Get("error saving slp send op return handler", err)
			}
		case memo.SlpTxTypeCommit:
			err := save.SlpCommit(info)
			if err != nil {
				return jerr.Get("error saving slp commit op return handler", err)
			}
		default:
			return jerr.Newf("unknown slp tx type (%s) op return handler", info.PushData[2])
		}
		return nil
	},
}
