package parse

import (
	wire2 "github.com/gcash/bchd/wire"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
)

func GetInfoGCash(tx *wire2.MsgTx) TxInfo {
	msgTx := wire.NewMsgTx(0)
	for _, in := range tx.TxIn {
		hash, _ := chainhash.NewHash(in.PreviousOutPoint.Hash.CloneBytes())
		newIn := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: in.PreviousOutPoint.Index,
		}, in.SignatureScript)
		msgTx.TxIn = append(msgTx.TxIn, newIn)
	}
	for _, out := range tx.TxOut {
		msgTx.TxOut = append(msgTx.TxOut, wire.NewTxOut(out.Value, out.PkScript))
	}
	return GetTxInfoMsg(msgTx)
}
