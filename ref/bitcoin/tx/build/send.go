package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/wallet"
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
		return nil, jerr.New(wallet.UnknownAddressTypeErrorMessage)
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
		return nil, jerr.Get("error building send tx", err)
	}
	return tx, nil
}
