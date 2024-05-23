package build

import (
	"fmt"
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
		return nil, fmt.Errorf("error building link account accept tx; %w", err)
	}
	return tx, nil
}
