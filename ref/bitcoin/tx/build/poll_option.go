package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

type PollOptionRequest struct {
	Wallet     Wallet
	PollTxHash []byte
	Option     string
}

func PollOption(request PollOptionRequest) (*memo.Tx, error) {
	tx, err := SimpleSingle(request.Wallet, []*memo.Output{{
		Script: &script.PollOption{
			PollTxHash: request.PollTxHash,
			Option:     request.Option,
		},
	}})
	if err != nil {
		return nil, jerr.Get("error building poll create tx", err)
	}
	return tx, nil
}
