package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
)

type LinkRevokeRequest struct {
	Wallet       Wallet
	AcceptTxHash []byte
	Message      string
}

func LinkRevoke(request LinkRevokeRequest) (*memo.Tx, error) {
	tx, err := gen.Tx(gen.TxRequest{
		Outputs: []*memo.Output{{
			Script: &script.LinkRevoke{
				AcceptTxHash: request.AcceptTxHash,
				Message:      request.Message,
			},
		}},
		Getter:  request.Wallet.Getter,
		Change:  request.Wallet.GetChange(),
		KeyRing: request.Wallet.KeyRing,
	})
	if err != nil {
		return nil, jerr.Get("error building link revoke tx", err)
	}
	return tx, nil
}
