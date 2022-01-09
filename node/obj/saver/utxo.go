package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
)

type Utxo struct {
	Verbose bool
}

func (u *Utxo) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var lockUtxos []*item.LockUtxo
	var txOutputs []*item.TxOutput
	var txInputs []*item.TxInput
	var ins []memo.Out
	var lockHashes [][]byte
	for _, msgTx := range block.Transactions {
		txHash := msgTx.TxHash()
		txHashBytes := txHash.CloneBytes()
		meta := parse.GetMeta(msgTx)
		if meta.Multi {
			jlog.Logf("FOUND meta with multi OP_RETURNS! %s\n", txHash)
		}
		specialIndexes := meta.GetSpecialIndexes()
		for g := range msgTx.TxIn {
			ins = append(ins, memo.Out{
				TxHash: msgTx.TxIn[g].PreviousOutPoint.Hash.CloneBytes(),
				Index:  msgTx.TxIn[g].PreviousOutPoint.Index,
			})
			txInputs = append(txInputs, &item.TxInput{
				TxHash:    txHashBytes,
				Index:     uint32(g),
				PrevHash:  msgTx.TxIn[g].PreviousOutPoint.Hash.CloneBytes(),
				PrevIndex: msgTx.TxIn[g].PreviousOutPoint.Index,
			})
		}
		for g, txOut := range msgTx.TxOut {
			var lockUtxo = &item.LockUtxo{
				Hash:     txHashBytes,
				Index:    uint32(g),
				Value:    txOut.Value,
				LockHash: script.GetLockHash(txOut.PkScript),
			}
			if jutil.InUintSlice(uint(g), specialIndexes) {
				lockUtxo.Special = true
			}
			lockUtxos = append(lockUtxos, lockUtxo)
			lockHashes = append(lockHashes, lockUtxo.LockHash)
			var txOutput = &item.TxOutput{
				TxHash:   lockUtxo.Hash,
				Index:    lockUtxo.Index,
				Value:    lockUtxo.Value,
				LockHash: lockUtxo.LockHash,
			}
			txOutputs = append(txOutputs, txOutput)
		}
	}
	var outs []memo.Out
LockUtxoLoop:
	for i := 0; i < len(lockUtxos); i++ {
		lockUtxo := lockUtxos[i]
		for _, in := range ins {
			if bytes.Equal(in.TxHash, lockUtxo.Hash) && in.Index == lockUtxo.Index {
				lockUtxos = append(lockUtxos[:i], lockUtxos[i+1:]...)
				i--
				continue LockUtxoLoop
			}
		}
		outs = append(outs, memo.Out{
			TxHash: lockUtxos[i].Hash,
			Index:  lockUtxos[i].Index,
		})
	}
	outputInputs, err := item.GetOutputInputs(outs)
	if err != nil {
		return jerr.Get("error getting utxo output inputs", err)
	}
	for i := 0; i < len(lockUtxos); i++ {
		lockUtxo := lockUtxos[i]
		for g, outputInput := range outputInputs {
			if bytes.Equal(outputInput.PrevHash, lockUtxo.Hash) &&
				outputInput.PrevIndex == lockUtxo.Index {
				lockUtxos = append(lockUtxos[:i], lockUtxos[i+1:]...)
				i--
				outputInputs = append(outputInputs[:g], outputInputs[g+1:]...)
				break
			}
		}
	}
	var numLockUtxos = len(lockUtxos)
	var numLockUtxosAndTxOutputs = numLockUtxos + len(txOutputs)
	var objects = make([]item.Object, numLockUtxosAndTxOutputs+len(txInputs))
	for i := range lockUtxos {
		objects[i] = lockUtxos[i]
	}
	for i := range txOutputs {
		objects[numLockUtxos+i] = txOutputs[i]
	}
	for i := range txInputs {
		objects[numLockUtxosAndTxOutputs+i] = txInputs[i]
	}
	if err = item.Save(objects); err != nil {
		return jerr.Get("error saving new lock utxos", err)
	}
	matchingTxOutputs, err := item.GetTxOutputs(ins)
	if err != nil {
		return jerr.Get("error getting matching tx outputs for inputs", err)
	}
	var spentOuts = make([]*item.LockUtxo, len(matchingTxOutputs))
	for i := range matchingTxOutputs {
		spentOuts[i] = &item.LockUtxo{
			LockHash: matchingTxOutputs[i].LockHash,
			Hash:     matchingTxOutputs[i].TxHash,
			Index:    matchingTxOutputs[i].Index,
		}
		lockHashes = append(lockHashes, matchingTxOutputs[i].LockHash)
	}
	if err = item.RemoveLockUtxos(spentOuts); err != nil {
		return jerr.Get("error removing lock utxos", err)
	}
	if err = item.RemoveLockBalances(lockHashes); err != nil {
		return jerr.Get("error removing lock balances", err)
	}
	return nil
}

func NewUtxo(verbose bool) *Utxo {
	return &Utxo{
		Verbose: verbose,
	}
}
