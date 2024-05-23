package gen

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

func Tx(request TxRequest) (*memo.Tx, error) {
	create := Create{
		Request:     request,
		InputsToUse: request.InputsToUse,
		Outputs:     request.Outputs,
	}
	msgTx, err := create.Build()
	if err != nil {
		return nil, fmt.Errorf("error building tx in generator; %w", err)
	}
	err = create.Sign(msgTx, request.KeyRing)
	if err != nil {
		return nil, fmt.Errorf("error signing tx in generator; %w", err)
	}
	memoTx := GetMemoTx(msgTx, create.InputsToUse, create.Outputs)
	err = create.markSpentAndChangeForGetter(memoTx)
	if err != nil {
		return nil, fmt.Errorf("error marking spent and change for getter in generator; %w", err)
	}
	return memoTx, nil
}

func (c *Create) markSpentAndChangeForGetter(memoTx *memo.Tx) error {
	var getter = c.Request.Getter
	if getter == nil {
		return nil
	}
	getter.MarkUTXOsUsed(c.InputsToUse)
	utxos := script.GetOutputUTXOs(memoTx)
	if len(utxos) != len(memoTx.Outputs) {
		return fmt.Errorf("error marking change: utxo count (%d) != output count (%d)",
			len(utxos), len(memoTx.Outputs))
	}
	for i, output := range memoTx.Outputs {
		if output.Amount != 0 {
			getter.AddChangeUTXO(utxos[i])
		}
	}
	return nil
}
