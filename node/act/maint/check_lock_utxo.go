package maint

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/act/block_tx"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"sort"
)

type CheckLockUtxo struct {
	MissingUtxos   []memo.Out
	CheckedOutputs int
	FoundInputs    int
	FoundUtxos     int
	OutsRemoved1   int
	OutsRemoved2   int
}

func (c *CheckLockUtxo) Check(blockHash []byte) error {
	if err := block_tx.NewLoopRaw(func(blockTxesRaw []*item.BlockTxRaw) error {
		var outs []memo.Out
		for i := range blockTxesRaw {
			msg, err := memo.GetMsgFromRaw(blockTxesRaw[i].Raw)
			if err != nil {
				return jerr.Get("error getting tx from raw block tx", err)
			}
			txHash := msg.TxHash()
			txHashBytes := txHash.CloneBytes()
			for i, txOut := range msg.TxOut {
				if txOut.Value == 0 {
					continue
				}
				outs = append(outs, memo.Out{
					TxHash:   txHashBytes,
					Index:    uint32(i),
					LockHash: script.GetLockHash(txOut.PkScript),
				})
			}
		}
		lenOuts := len(outs)
		sort.Slice(outs, func(i, j int) bool {
			if !bytes.Equal(outs[i].TxHash, outs[j].TxHash) {
				return jutil.ByteLT(outs[i].TxHash, outs[j].TxHash)
			}
			return outs[i].Index < outs[j].Index
		})
		outputInputs, err := chain.GetOutputInputs(outs)
		if err != nil {
			return jerr.Get("error getting output inputs for check lock utxos", err)
		}
		c.FoundInputs += len(outputInputs)
		sort.Slice(outputInputs, func(i, j int) bool {
			if outputInputs[i].PrevHash != outputInputs[j].PrevHash {
				return jutil.ByteLT(outputInputs[i].PrevHash[:], outputInputs[j].PrevHash[:])
			}
			return outputInputs[i].PrevIndex < outputInputs[j].PrevIndex
		})
		var outIndex int
		for _, outputInput := range outputInputs {
			for ; outIndex < len(outs); outIndex++ {
				if bytes.Equal(outputInput.PrevHash[:], outs[outIndex].TxHash) &&
					outputInput.PrevIndex == outs[outIndex].Index {
					outs = append(outs[:outIndex], outs[outIndex+1:]...)
					outIndex--
					c.OutsRemoved1++
				} else if jutil.ByteLT(outputInput.PrevHash[:], outs[outIndex].TxHash) ||
					(bytes.Equal(outputInput.PrevHash[:], outs[outIndex].TxHash) &&
						outputInput.PrevIndex < outs[outIndex].Index) {
					break
				}
			}
		}
		lockUtxos, err := item.GetLockUtxosByOuts(outs)
		if err != nil {
			return jerr.Get("error getting lock utxos for check lock utxos", err)
		}
		c.FoundUtxos += len(lockUtxos)
		sort.Slice(lockUtxos, func(i, j int) bool {
			if !bytes.Equal(lockUtxos[i].Hash, lockUtxos[j].Hash) {
				return jutil.ByteLT(lockUtxos[i].Hash, lockUtxos[j].Hash)
			}
			return lockUtxos[i].Index < lockUtxos[j].Index
		})
		outIndex = 0
		for _, lockUtxo := range lockUtxos {
			for ; outIndex < len(outs); outIndex++ {
				if bytes.Equal(lockUtxo.Hash, outs[outIndex].TxHash) &&
					lockUtxo.Index == outs[outIndex].Index {
					outs = append(outs[:outIndex], outs[outIndex+1:]...)
					outIndex--
					c.OutsRemoved2++
				} else if jutil.ByteLT(lockUtxo.Hash, outs[outIndex].TxHash) ||
					(bytes.Equal(lockUtxo.Hash, outs[outIndex].TxHash) &&
						lockUtxo.Index < outs[outIndex].Index) {
					break
				}
			}
		}
		c.MissingUtxos = append(c.MissingUtxos, outs...)
		c.CheckedOutputs += lenOuts
		return nil
	}).Process(blockHash); err != nil {
		jerr.Get("fatal error processing block txs for check lock utxo", err).Fatal()
	}
	return nil
}

func NewCheckLockUtxo() *CheckLockUtxo {
	return &CheckLockUtxo{}
}
