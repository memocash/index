package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type LinkRequestRequest struct {
	OldWallet    Wallet
	NewWallet    Wallet
	ParentPkHash []byte
	Message      string
}

func LinkRequest(request LinkRequestRequest) (*memo.Tx, error) {
	if len(request.ParentPkHash) == 0 {
		request.ParentPkHash = request.NewWallet.GetPkHash()
	}
	tx, err := gen.Tx(gen.TxRequest{
		Outputs: []*memo.Output{{
			Script: &script.LinkRequest{
				ParentPkHash: request.ParentPkHash,
				Message:      request.Message,
			},
		}},
		Getter:  request.OldWallet.Getter,
		Change:  request.NewWallet.GetChange(),
		KeyRing: request.OldWallet.KeyRing,
	})
	if err != nil {
		return nil, fmt.Errorf("error building link account request tx; %w", err)
	}
	return tx, nil
}
