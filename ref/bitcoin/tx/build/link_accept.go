package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type LinkAcceptRequest struct {
	Wallet        Wallet
	RequestTxHash []byte
	Message       string
}

func LinkAccept(request LinkAcceptRequest) (*memo.Tx, error) {
	tx, err := gen.Tx(gen.TxRequest{
		Outputs: []*memo.Output{{
			Script: &script.LinkAccept{
				RequestTxHash: request.RequestTxHash,
				Message:       request.Message,
			},
		}},
		Getter:  request.Wallet.Getter,
		Change:  request.Wallet.GetChange(),
		KeyRing: request.Wallet.KeyRing,
	})
	if err != nil {
		return nil, jerr.Get("error building link account accept tx", err)
	}
	return tx, nil
}
