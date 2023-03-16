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

func AttachToTxs(preloads []string, txs []*model.Tx) error {
	if jutil.StringsInSlice([]string{"inputs", "raw"}, preloads) {
		if err := attachInputsToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringsInSlice([]string{"outputs", "raw"}, preloads) {
		if err := attachOutputsToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringsInSlice([]string{"version", "locktime", "raw"}, preloads) {
		if err := attachInfoToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringInSlice("raw", preloads) {
		if err := attachRawsToTxs(txs); err != nil {
			return err
		}
	}
	if jutil.StringInSlice("seen", preloads) {
		if err := attachSeensToTxs(txs); err != nil {
			return err
		}
	}
	var allOutputs []*model.TxOutput
	for _, tx := range txs {
		allOutputs = append(allOutputs, tx.Outputs...)
	}
	if err := AttachToOutputs(GetPrefixPreloads(preloads, "outputs."), allOutputs); err != nil {
		return err
	}
	return nil
}

func attachInputsToTxs(txs []*model.Tx) error {
	var txHashes = make([][32]byte, len(txs))
	for i := range txs {
		txHashes[i] = txs[i].Hash
	}
	txInputs, err := chain.GetTxInputsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting tx inputs for model tx; %w", err)
	}
	for i := range txs {
		for j := range txInputs {
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
		}
		sort.Slice(txs[i].Inputs, func(a, b int) bool {
			return txs[i].Inputs[a].Index < txs[i].Inputs[b].Index
		})
	}
	return nil
}

func attachOutputsToTxs(txs []*model.Tx) error {
	var txHashes = make([][32]byte, len(txs))
	for i := range txs {
		txHashes[i] = txs[i].Hash
	}
	txOutputs, err := chain.GetTxOutputsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting tx outputs for model tx; %w", err)
	}
	for i := range txs {
		for j := range txOutputs {
			if txs[i].Hash != txOutputs[j].TxHash {
				continue
			}
			txs[i].Outputs = append(txs[i].Outputs, &model.TxOutput{
				Hash:   txOutputs[j].TxHash,
				Index:  txOutputs[j].Index,
				Amount: txOutputs[j].Value,
				Script: txOutputs[j].LockScript,
			})
		}
		sort.Slice(txs[i].Outputs, func(a, b int) bool {
			return txs[i].Outputs[a].Index < txs[i].Outputs[b].Index
		})
	}
	return nil
}

func attachInfoToTxs(txs []*model.Tx) error {
	var txHashes [][32]byte
	for i := range txs {
		if txs[i].Version == 0 {
			txHashes = append(txHashes, txs[i].Hash)
		}
	}
	if len(txHashes) == 0 {
		return nil
	}
	chainTxs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		return fmt.Errorf("error getting chain txs for raw; %w", err)
	}
	for i := range txs {
		for j := range chainTxs {
			if txs[i].Hash != chainTxs[j].TxHash {
				continue
			}
			txs[i].Version = chainTxs[j].Version
			txs[i].LockTime = chainTxs[j].LockTime
			break
		}
	}
	return nil
}

func attachRawsToTxs(txs []*model.Tx) error {
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

func attachSeensToTxs(txs []*model.Tx) error {
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
		for j := range txSeens {
			if txs[i].Hash != txSeens[j].TxHash {
				continue
			}
			txs[i].Seen = model.Date(txSeens[j].Timestamp)
			break
		}
	}
	return nil
}
