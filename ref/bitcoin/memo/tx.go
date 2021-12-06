package memo

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

type TxInput struct {
	PkScript     []byte
	PkHash       []byte
	Value        int64
	PrevOutHash  []byte
	PrevOutIndex uint32
}

func (t TxInput) GetHashIndexString() string {
	return hs.GetHashIndexString(t.PrevOutHash, t.PrevOutIndex)
}

func (t TxInput) GetDisasmString() string {
	s, _ := txscript.DisasmString(t.PkScript)
	return s
}

type Tx struct {
	SelfPkHash  []byte
	MsgTx       *wire.MsgTx
	Inputs      []*TxInput
	Outputs     []*Output
	OpReturn    *Output
	AncestorsNC uint
}

func (tx Tx) GetHash() []byte {
	txHash := tx.MsgTx.TxHash()
	return txHash.CloneBytes()
}

func (tx Tx) GetType() OutputType {
	if tx.OpReturn != nil {
		return tx.OpReturn.GetType()
	}
	return OutputTypeUnknown
}
