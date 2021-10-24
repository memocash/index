package gen

import (
	"github.com/memocash/server/ref/bitcoin/memo"
)

type Create struct {
	Request         TxRequest
	PotentialInputs []memo.UTXO
	InputsToUse     []memo.UTXO
	Outputs         []*memo.Output
}
