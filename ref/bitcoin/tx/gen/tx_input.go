package gen

import "github.com/memocash/server/ref/bitcoin/memo"

func GetNonPointerTxInputs(pointerTxInputs []*memo.TxInput) []memo.TxInput {
	var inputs = make([]memo.TxInput, len(pointerTxInputs))
	for i := range pointerTxInputs {
		inputs[i] = *pointerTxInputs[i]
	}
	return inputs
}
