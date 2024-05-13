package tx_raw

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/ref/bitcoin/memo"
	"sort"
)

type TxRaw struct {
	Hash [32]byte
	Raw  []byte
}

func GetSingle(ctx context.Context, txHash [32]byte) (*TxRaw, error) {
	raws, err := Get(ctx, [][32]byte{txHash})
	if err != nil {
		return nil, jerr.Get("error getting tx raws for single", err)
	} else if len(raws) != 1 {
		return nil, jerr.Newf("error tx raw not found for single", err)
	}
	return raws[0], nil
}

func Get(ctx context.Context, txHashes [][32]byte) ([]*TxRaw, error) {
	sort.Slice(txHashes, func(i, j int) bool {
		return jutil.ByteLT(txHashes[i][:], txHashes[j][:])
	})
	txs, err := chain.GetTxsByHashes(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for raw", err)
	}
	txInputs, err := chain.GetTxInputsByHashes(ctx, txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for raw", err)
	}
	sort.Slice(txInputs, func(i, j int) bool {
		return txInputs[i].Index < txInputs[j].Index
	})
	txOutputs, err := chain.GetTxOutputsByHashes(ctx, txHashes)
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
		var index uint32 = 0
		for _, txIn := range txInputs {
			if txIn.TxHash != tx.TxHash {
				continue
			}
			if txIn.Index != index {
				return nil, jerr.Newf("tx input index missing: %s %d %d",
					chainhash.Hash(txIn.TxHash), txIn.Index, index)
			}
			index++
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
		index = 0
		for _, txOut := range txOutputs {
			if txOut.TxHash != tx.TxHash {
				continue
			}
			if txOut.Index != index {
				return nil, jerr.Newf("tx output index missing: %s %d %d",
					chainhash.Hash(txOut.TxHash), txOut.Index, index)
			}
			index++
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
