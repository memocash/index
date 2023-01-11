package tx_raw

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"sort"
)

type TxRaw struct {
	Hash [32]byte
	Raw  []byte
}

func GetSingle(txHash [32]byte) (*TxRaw, error) {
	raws, err := Get([][32]byte{txHash})
	if err != nil {
		return nil, jerr.Get("error getting tx raws for single", err)
	} else if len(raws) != 1 {
		return nil, jerr.Newf("error tx raw not found for single", err)
	}
	return raws[0], nil
}

func Get(txHashes [][32]byte) ([]*TxRaw, error) {
	txs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for raw", err)
	}
	txInputs, err := chain.GetTxInputsByHashes(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for raw", err)
	}
	sort.Slice(txInputs, func(i, j int) bool {
		return txInputs[i].Index < txInputs[j].Index
	})
	txOutputs, err := chain.GetTxOutputsByHashes(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx outputs for raw", err)
	}
	sort.Slice(txOutputs, func(i, j int) bool {
		return txOutputs[i].Index < txOutputs[j].Index
	})
	var txRaws []*TxRaw
	for _, tx := range txs {
		var msgTx = &wire.MsgTx{
			Version:  tx.Version,
			LockTime: tx.LockTime,
		}
		for i, txIn := range txInputs {
			if txIn.TxHash != tx.TxHash {
				continue
			}
			if txIn.Index != uint32(i) {
				return nil, jerr.Newf("tx input index missing: %d %d", txIn.Index, i)
			}
			msgTx.TxIn = append(msgTx.TxIn, &wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					Hash:  txIn.PrevHash,
					Index: txIn.PrevIndex,
				},
				SignatureScript: txIn.UnlockScript,
				Sequence:        txIn.Sequence,
			})
		}
		if len(msgTx.TxIn) == 0 {
			return nil, jerr.Newf("tx inputs missing for tx: %s", chainhash.Hash(tx.TxHash))
		}
		for i, txOut := range txOutputs {
			if txOut.TxHash != tx.TxHash {
				continue
			}
			if txOut.Index != uint32(i) {
				return nil, jerr.Newf("tx output index missing: %d %d", txOut.Index, i)
			}
			msgTx.TxOut = append(msgTx.TxOut, &wire.TxOut{
				Value:    txOut.Value,
				PkScript: txOut.LockScript,
			})
		}
		if len(msgTx.TxOut) == 0 {
			return nil, jerr.Newf("tx outputs missing for tx: %s", chainhash.Hash(tx.TxHash))
		}
		if msgTx.TxHash() != tx.TxHash {
			return nil, jerr.Newf("tx hash mismatch for raw: %s %s",
				msgTx.TxHash(), chainhash.Hash(tx.TxHash))
		}
		txRaws = append(txRaws, &TxRaw{
			Hash: tx.TxHash,
			Raw:  memo.GetRaw(msgTx),
		})
	}
	return txRaws, nil
}
