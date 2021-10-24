package gen

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/server/ref/bitcoin/memo"
)

func getMemoTx(tx *wire.MsgTx, utxos []memo.UTXO, outputs []*memo.Output) *memo.Tx {
	var memoInputs []*memo.TxInput
	var maxAncestors uint
	for i, utxo := range utxos {
		memoInputs = append(memoInputs, &utxos[i].Input)
		if utxo.AncestorsNC > maxAncestors {
			maxAncestors = utxo.AncestorsNC
		}
	}
	var spendOutput *memo.Output
	for _, output := range outputs {
		if output.GetType() != memo.OutputTypeP2PKH {
			spendOutput = output
		}
	}
	return &memo.Tx{
		MsgTx:       tx,
		Inputs:      memoInputs,
		Outputs:     outputs,
		OpReturn:    spendOutput,
		AncestorsNC: maxAncestors + 1,
	}
}
