package memo

import (
	"bytes"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
)

type UTXO struct {
	Input       TxInput
	SlpToken    []byte
	SlpQuantity uint64
	SlpType     SlpType
	SellTxHash  []byte
	AncestorsNC uint // Ancestors not confirmed
}

func (u UTXO) IsEqual(u2 UTXO) bool {
	return bytes.Equal(u.Input.PrevOutHash, u2.Input.PrevOutHash) && u.Input.PrevOutIndex == u2.Input.PrevOutIndex
}

func (u UTXO) IsPrevOutSet() bool {
	return len(u.Input.PrevOutHash) > 0
}

func (u UTXO) IsSlp() bool {
	return len(u.SlpToken) > 0
}

func (u UTXO) IsSellTokenInput() bool {
	return len(u.SellTxHash) > 0
}

func (u UTXO) AtAncestorLimit() bool {
	return u.AncestorsNC >= MaxAncestors
}

type UTXORequest struct {
	TokenHash []byte
	Baton     bool
}

func (r UTXORequest) GetHashString() string {
	return hs.GetTxString(r.TokenHash)
}

func GetNonPointerUtxos(utxos []*UTXO) []UTXO {
	var nonPointerUtxos = make([]UTXO, len(utxos))
	for i := range utxos {
		nonPointerUtxos[i] = *utxos[i]
	}
	return nonPointerUtxos
}
