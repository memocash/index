package script

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

func len64(b []byte) int64 {
	return int64(len(b))
}

func Len64(b []byte) int64 {
	return len64(b)
}

func GetOutputUTXOs(tx *memo.Tx) []memo.UTXO {
	utxos := convertTxInputsToUTXOs(GetOutputs(tx))
	for i := range utxos {
		utxos[i].AncestorsNC = tx.AncestorsNC
	}
	if tx.OpReturn == nil {
		return utxos
	}
	switch s := tx.OpReturn.Script.(type) {
	case *TokenSend:
		for i, quantity := range s.Quantities {
			if quantity > 0 {
				if len(utxos) > i+1 {
					utxos[i+1].SlpType = memo.SlpTxTypeSend
					utxos[i+1].SlpQuantity = quantity
					utxos[i+1].SlpToken = s.TokenHash
				}
			}
		}
	case *TokenCreate:
		const createIndex = 1
		if len(utxos) > createIndex {
			utxos[createIndex].SlpType = memo.SlpTxTypeGenesis
			utxos[createIndex].SlpQuantity = s.Quantity
			utxos[createIndex].SlpToken = tx.GetHash()
		}
	case *TokenMint:
		const createIndex = 1
		if len(utxos) > createIndex {
			utxos[createIndex].SlpType = memo.SlpTxTypeMint
			utxos[createIndex].SlpQuantity = s.Quantity
			utxos[createIndex].SlpToken = s.TokenHash
		}
	}
	return utxos
}

func convertTxInputsToUTXOs(txInputs []*memo.TxInput) []memo.UTXO {
	var utxos []memo.UTXO
	for _, txInput := range txInputs {
		utxos = append(utxos, memo.UTXO{
			Input: *txInput,
		})
	}
	return utxos
}

func GetAddress(script memo.Script) wallet.Address {
	switch v := script.(type) {
	case *P2pkh:
		return wallet.GetAddressFromPkHash(v.PkHash)
	case *P2sh:
		return wallet.GetAddressFromScriptHash(v.ScriptHash)
	}
	return wallet.Address{}
}

func GetOutputs(tx *memo.Tx) []*memo.TxInput {
	txHash := tx.MsgTx.TxHash()
	var txInputs []*memo.TxInput
	for index, out := range tx.MsgTx.TxOut {
		txInputs = append(txInputs, &memo.TxInput{
			PrevOutHash:  txHash.CloneBytes(),
			PrevOutIndex: uint32(index),
			Value:        out.Value,
			PkScript:     out.PkScript,
			PkHash:       GetAddress(tx.Outputs[index].Script).GetPkHash(),
		})
	}
	return txInputs
}

func IsBigMemo(outputs []*memo.Output) bool {
	if len(outputs) == 0 {
		return false
	}
	switch v := outputs[0].Script.(type) {
	case *Post:
		if len(v.Message) > memo.OldMaxPostSize {
			return true
		}
	case *TopicMessage:
		if (len(v.TopicName) + len(v.Message)) > memo.OldMaxTagMessageSize {
			return true
		}
	case *Reply:
		if len(v.Message) > memo.OldMaxReplySize {
			return true
		}
	case *Save:
		return true
	}
	return false
}

func GetLockHash(pkScript []byte) []byte {
	if txscript.IsPubKey(pkScript) {
		pushedData, _ := txscript.PushedData(pkScript)
		if len(pushedData) == 1 {
			address := wallet.GetAddress(pushedData[0])
			pkHashScript := GetLockHashForAddress(address)
			if len(pkHashScript) == 0 {
				return jutil.GetSha256Hash(pkHashScript)
			}
		}
	}
	return jutil.GetSha256Hash(pkScript)
}

func GetLockHashForAddress(address wallet.Address) []byte {
	pkHashScript, err := P2pkh{PkHash: address.GetPkHash()}.Get()
	if err == nil {
		return jutil.GetSha256Hash(pkHashScript)
	}
	return nil
}
