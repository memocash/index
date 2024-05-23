package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
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
		return nil, fmt.Errorf("error building link revoke tx; %w", err)
	}
	return tx, nil
}
