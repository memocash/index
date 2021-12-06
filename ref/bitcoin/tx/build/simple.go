package build

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
)

func Simple(w Wallet, outputs []*memo.Output) ([]*memo.Tx, error) {
	return SimpleWithInputs(w, outputs, nil)
}

func SimpleWithInputs(w Wallet, outputs []*memo.Output, inputs []memo.UTXO) ([]*memo.Tx, error) {
	txs, err := gen.Multi(gen.MultiRequest{
		Outputs:      outputs,
		Getter:       w.Getter,
		FaucetGetter: w.FaucetGetter,
		FaucetSaver:  w.FaucetSaver,
		Change:       w.GetChange(),
		KeyRing:      w.KeyRing,
		InputsToUse:  inputs,
	})
	if err != nil {
		return nil, jerr.Get("error building simple tx", err)
	}
	return txs, nil
}

func SimpleSingle(w Wallet, outputs []*memo.Output) (*memo.Tx, error) {
	tx, err := gen.Tx(gen.TxRequest{
		Outputs: outputs,
		Getter:  w.Getter,
		Change:  w.GetChange(),
		KeyRing: w.KeyRing,
	})
	if err != nil {
		return nil, jerr.Get("error building simple tx", err)
	}
	return tx, nil
}
