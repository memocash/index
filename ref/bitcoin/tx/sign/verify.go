package sign

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
)

type Output struct {
	PkScript []byte
	Value    int64
}

func Verify(tx *wire.MsgTx, inputTxs []*wire.MsgTx) error {
	var outputs = make([]*Output, len(tx.TxIn))
	for _, inputTx := range inputTxs {
		inputTxHash := inputTx.TxHash()
		for i, txIn := range tx.TxIn {
			if txIn.PreviousOutPoint.Hash.IsEqual(&inputTxHash) {
				index := int(txIn.PreviousOutPoint.Index)
				if len(inputTx.TxOut) <= index {
					return jerr.Newf("error tx found but output index too high (%d %d)", len(inputTx.TxOut),
						index)
				}
				outputs[i] = &Output{
					PkScript: inputTx.TxOut[index].PkScript,
					Value:    inputTx.TxOut[index].Value,
				}
			}
		}
	}
	if err := VerifyWithOutputs(tx, outputs); err != nil {
		return jerr.Get("error verifying with outputs", err)
	}
	return nil
}

func VerifyWithOutputs(tx *wire.MsgTx, outputs []*Output) error {
	if len(tx.TxIn) != len(outputs) {
		return jerr.Newf("error tx input length does not match outputs length (%d %d)", len(tx.TxIn), len(outputs))
	}
	flags := txscript.StandardVerifyFlags
	for index := range tx.TxIn {
		out := outputs[index]
		vm, err := txscript.NewEngine(out.PkScript, tx, index, flags, nil, out.Value)
		if err != nil {
			return jerr.Getf(err, "error getting new tx script engine for input: %s:%d", tx.TxHash(), index)
		}
		if err = vm.Execute(); err != nil {
			return jerr.Getf(err, "error executing vm engine for input: %s:%d", tx.TxHash(), index)
		}
	}
	return nil
}

func VerifySignature(pkScript []byte, spendTx *wire.MsgTx, index int, amount int64) error {
	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(pkScript, spendTx, index, flags, nil, amount)
	if err != nil {
		return jerr.Get("error getting new tx script engine", err)
	}
	if err = vm.Execute(); err != nil {
		return jerr.Get("error executing vm engine", err)
	}
	return nil
}
