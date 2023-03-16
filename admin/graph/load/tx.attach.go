package load

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"sort"
	"time"
)

func AttachAllToTxs(preloads []string, txs []*model.Tx) error {
	if jutil.StringsInSlice([]string{"inputs", "raw"}, preloads) {
		if err := AttachInputsToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringsInSlice([]string{"outputs", "raw"}, preloads) {
		if err := AttachOutputsToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringsInSlice([]string{"version", "locktime", "raw"}, preloads) {
		if err := AttachInfoToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringInSlice("raw", preloads) {
		if err := AttachRawsToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringInSlice("seen", preloads) {
		if err := AttachSeensToTxs(txs); err != nil {
			return err
		}
	}
	return nil
}

func AttachInputsToTxs(txs []*model.Tx) error {
	var txHashes = make([][32]byte, len(txs))
	for i := range txs {
		txHashes[i] = txs[i].Hash
	}
	txInputs, err := chain.GetTxInputsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting tx inputs for model tx; %w", err)
	}
	for i := range txs {
		for j := 0; j < len(txInputs); j++ {
			if txs[i].Hash != txInputs[j].TxHash {
				continue
			}
			txs[i].Inputs = append(txs[i].Inputs, &model.TxInput{
				Hash:      txInputs[j].TxHash,
				Index:     txInputs[j].Index,
				PrevHash:  txInputs[j].PrevHash,
				PrevIndex: txInputs[j].PrevIndex,
				Sequence:  txInputs[j].Sequence,
				Script:    txInputs[j].UnlockScript,
			})
			txInputs = append(txInputs[:j], txInputs[j+1:]...)
			j--
		}
		sort.Slice(txs[i].Inputs, func(a, b int) bool {
			return txs[i].Inputs[a].Index < txs[i].Inputs[b].Index
		})
	}
	return nil
}

func AttachOutputsToTxs(txs []*model.Tx) error {
	var txHashes = make([][32]byte, len(txs))
	for i := range txs {
		txHashes[i] = txs[i].Hash
	}
	txOutputs, err := chain.GetTxOutputsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting tx outputs for model tx; %w", err)
	}
	for i := range txs {
		for j := 0; j < len(txOutputs); j++ {
			if txs[i].Hash != txOutputs[j].TxHash {
				continue
			}
			txs[i].Outputs = append(txs[i].Outputs, &model.TxOutput{
				Hash:   txOutputs[j].TxHash,
				Index:  txOutputs[j].Index,
				Amount: txOutputs[j].Value,
				Script: txOutputs[j].LockScript,
			})
			txOutputs = append(txOutputs[:j], txOutputs[j+1:]...)
			j--
		}
		sort.Slice(txs[i].Outputs, func(a, b int) bool {
			return txs[i].Outputs[a].Index < txs[i].Outputs[b].Index
		})
	}
	return nil
}

func AttachInfoToTxs(txs []*model.Tx) error {
	var txHashes = make([][32]byte, len(txs))
	for i := range txs {
		txHashes[i] = txs[i].Hash
	}
	chainTxs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting chain txs for raw; %w", err)
	}
	for i := range txs {
		for j := 0; j < len(chainTxs); j++ {
			if txs[i].Hash != chainTxs[j].TxHash {
				continue
			}
			txs[i].Version = chainTxs[j].Version
			txs[i].LockTime = chainTxs[j].LockTime
			j--
			break
		}
	}
	return nil
}

func AttachRawsToTxs(txs []*model.Tx) error {
	for i := range txs {
		var msgTx = &wire.MsgTx{
			Version:  txs[i].Version,
			LockTime: txs[i].LockTime,
		}
		for _, txIn := range txs[i].Inputs {
			msgTx.TxIn = append(msgTx.TxIn, &wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash(txIn.PrevHash),
					Index: txIn.PrevIndex,
				},
				SignatureScript: txIn.Script,
				Sequence:        txIn.Sequence,
			})
		}
		for _, txOut := range txs[i].Outputs {
			msgTx.TxOut = append(msgTx.TxOut, &wire.TxOut{
				Value:    txOut.Amount,
				PkScript: txOut.Script,
			})
		}
		if msgTx.TxHash() != chainhash.Hash(txs[i].Hash) {
			return fmt.Errorf("tx hash mismatch for raw: %s %s", msgTx.TxHash(), chainhash.Hash(txs[i].Hash))
		}
		txs[i].Raw = memo.GetRaw(msgTx)
	}
	return nil
}

func AttachSeensToTxs(txs []*model.Tx) error {
	var txHashes [][32]byte
	for i := range txs {
		if jutil.IsTimeZero(time.Time(txs[i].Seen)) {
			txHashes = append(txHashes, txs[i].Hash)
		}
	}
	if len(txHashes) == 0 {
		return nil
	}
	txSeens, err := chain.GetTxSeens(txHashes)
	if err != nil {
		return fmt.Errorf("error getting chain txs for raw; %w", err)
	}
	for i := range txs {
		for j := 0; j < len(txSeens); j++ {
			if txs[i].Hash != txSeens[j].TxHash {
				continue
			}
			txs[i].Seen = model.Date(txSeens[j].Timestamp)
			j--
			break
		}
	}
	return nil
}
