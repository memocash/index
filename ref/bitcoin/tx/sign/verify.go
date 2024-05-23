package sign

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
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
					return fmt.Errorf("error tx found but output index too high (%d %d)", len(inputTx.TxOut),
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
		return fmt.Errorf("error verifying with outputs; %w", err)
	}
	return nil
}

func VerifyWithOutputs(tx *wire.MsgTx, outputs []*Output) error {
	if len(tx.TxIn) != len(outputs) {
		return fmt.Errorf("error tx input length does not match outputs length (%d %d)", len(tx.TxIn), len(outputs))
	}
	flags := txscript.StandardVerifyFlags
	for index := range tx.TxIn {
		out := outputs[index]
		vm, err := txscript.NewEngine(out.PkScript, tx, index, flags, nil, out.Value)
		if err != nil {
			return fmt.Errorf("error getting new tx script engine for input: %s:%d; %w", tx.TxHash(), index, err)
		}
		if err = vm.Execute(); err != nil {
			return fmt.Errorf("error executing vm engine for input: %s:%d; %w", tx.TxHash(), index, err)
		}
	}
	return nil
}

func VerifySignature(pkScript []byte, spendTx *wire.MsgTx, index int, amount int64) error {
	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(pkScript, spendTx, index, flags, nil, amount)
	if err != nil {
		return fmt.Errorf("error getting new tx script engine; %w", err)
	}
	if err = vm.Execute(); err != nil {
		return fmt.Errorf("error executing vm engine; %w", err)
	}
	return nil
}
