package gen

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func Sign(msg *wire.MsgTx, inputs []memo.TxInput, keyRing wallet.KeyRing) error {
	for i := 0; i < len(msg.TxIn); i++ {
		if len(msg.TxIn[i].SignatureScript) > 0 {
			continue
		}
		sig, err := InputSignature(msg, i, keyRing, inputs)
		if err != nil {
			return jerr.Get("error signing input signature", err)
		}
		privateKey := keyRing.GetKey(inputs[i].PkHash)
		if !privateKey.IsSet() {
			return jerr.Newf("error unable to get private key from key ring for input: %s",
				inputs[i].GetHashIndexString())
		}
		pkData := privateKey.GetPublicKey().GetSerialized()
		err = AttachSignatureToInput(msg.TxIn[i], sig, pkData)
		if err != nil {
			return jerr.Get("error attaching signature to input", err)
		}
	}
	return nil
}

func InputSignature(tx *wire.MsgTx, index int, keyRing wallet.KeyRing, spendOuts []memo.TxInput) ([]byte, error) {
	if len(spendOuts[index].PkScript) == 0 {
		return nil, jerr.Newf("error no pk script for input signature: %s", spendOuts[index].GetHashIndexString())
	} else if len(spendOuts[index].PkHash) == 0 {
		return nil, jerr.Newf("error no pk hash for input signature: %s", spendOuts[index].GetHashIndexString())
	}
	sig, err := txscript.RawTxInECDSASignature(
		tx, index, spendOuts[index].PkScript, txscript.SigHashAll|wallet.SigHashForkID,
		keyRing.GetKey(spendOuts[index].PkHash).GetBtcEcPrivateKey(), spendOuts[index].Value)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func InputSignatureSingle(tx *wire.MsgTx, index int, privateKey wallet.PrivateKey, prevOut memo.Out) ([]byte, error) {
	sig, err := txscript.RawTxInECDSASignature(
		tx, index, prevOut.PkScript, txscript.SigHashSingle|txscript.SigHashAnyOneCanPay|wallet.SigHashForkID,
		privateKey.GetBtcEcPrivateKey(), prevOut.Value)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func AttachSignatureToInput(in *wire.TxIn, sig []byte, pkData []byte) error {
	sigScript, err := txscript.NewScriptBuilder().AddData(sig).AddData(pkData).Script()
	if err != nil {
		return jerr.Get("error getting signature script", err)
	}
	in.SignatureScript = sigScript
	return nil
}
