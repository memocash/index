package resolver

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/node/act/tx_raw"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

var txInputOutputLoaderConfig = dataloader.TxOutputLoaderConfig{
	Wait:     2 * time.Millisecond,
	MaxBatch: 100,
	Fetch: func(keys []model.HashIndex) ([]*model.TxOutput, []error) {
		var txHashes = make([][]byte, len(keys))
		var outs = make([]memo.Out, len(keys))
		for i := range keys {
			hash, err := chainhash.NewHashFromStr(keys[i].Hash)
			if err != nil {
				return nil, []error{jerr.Get("error parsing spend tx hash for output", err)}
			}
			txHashes[i] = hash.CloneBytes()
			outs[i] = memo.Out{
				TxHash: hash.CloneBytes(),
				Index:  keys[i].Index,
			}
		}
		txRaws, err := tx_raw.Get(txHashes)
		if err != nil {
			return nil, []error{jerr.Get("error getting tx raws for input output loader", err)}
		}
		var outputs = make([]*model.TxOutput, len(outs))
		for i := range outs {
			outputHash, err := chainhash.NewHash(outs[i].TxHash)
			if err != nil {
				return nil, []error{jerr.Get("error getting input output hash", err)}
			}
			outputs[i] = &model.TxOutput{
				Hash:  outputHash.String(),
				Index: outs[i].Index,
			}
			for _, txRaw := range txRaws {
				if bytes.Equal(txRaw.Hash, outs[i].TxHash) {
					tx, err := memo.GetMsgFromRaw(txRaw.Raw)
					if err != nil {
						return nil, []error{jerr.Get("error getting message from raw for input output loader", err)}
					}
					if len(tx.TxOut) <= int(outs[i].Index) {
						return nil, []error{jerr.Newf("error tx outs not long enough for index: %d %d",
							len(tx.TxOut), outs[i].Index)}
					}
					outputs[i].Amount = tx.TxOut[outs[i].Index].Value
					outputs[i].Script = hex.EncodeToString(tx.TxOut[outs[i].Index].PkScript)
				}
			}
		}
		return outputs, nil
	},
}
