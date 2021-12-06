package gen

import (
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Create struct {
	Request         TxRequest
	PotentialInputs []memo.UTXO
	InputsToUse     []memo.UTXO
	Outputs         []*memo.Output
}
