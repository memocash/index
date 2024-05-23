package build

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type SendRequest struct {
	Wallet  Wallet
	Address wallet.Address
	Message string
	Amount  int64
}

func Send(request SendRequest) (*memo.Tx, error) {
	var outputs []*memo.Output
	output := gen.GetAddressOutput(request.Address, request.Amount)
	if output == nil {
		return nil, fmt.Errorf(wallet.UnknownAddressTypeErrorMessage)
	}
	outputs = append(outputs, output)
	if request.Message != "" {
		outputs = append([]*memo.Output{{
			Script: &script.Send{
				Hash:    request.Address.ScriptAddress(),
				Message: request.Message,
			},
		}}, outputs...)
	}
	tx, err := gen.Tx(gen.TxRequest{
		Getter:  request.Wallet.Getter,
		Outputs: outputs,
		Change:  request.Wallet.GetChange(),
		KeyRing: request.Wallet.KeyRing,
	})
	if err != nil {
		return nil, fmt.Errorf("error building send tx; %w", err)
	}
	return tx, nil
}
