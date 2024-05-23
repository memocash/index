package gen

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func TxUnsigned(request TxRequest) (*memo.Tx, error) {
	create := Create{
		Request:     request,
		InputsToUse: request.InputsToUse,
		Outputs:     request.Outputs,
	}
	msgTx, err := create.Build()
	if err != nil {
		return nil, fmt.Errorf("error building tx; %w", err)
	}
	memoTx := GetMemoTx(msgTx, create.InputsToUse, create.Outputs)
	return memoTx, nil
}
