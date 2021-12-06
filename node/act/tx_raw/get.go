package tx_raw

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
)

type TxRaw struct {
	Hash []byte
	Raw  []byte
}

func Get(txHashes [][]byte) ([]*TxRaw, error) {
	txBlocks, err := item.GetTxBlocks(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx blocks for double spend lock hashes", err)
	}
	var mempoolTxHashes [][]byte
Loop:
	for _, txHash := range txHashes {
		for _, txBlock := range txBlocks {
			if bytes.Equal(txBlock.TxHash, txHash) {
				continue Loop
			}
		}
		mempoolTxHashes = append(mempoolTxHashes, txHash)
	}
	txBlockRaws, err := item.GetRawTxBlocksByHashes(txBlocks)
	if err != nil {
		return nil, jerr.Get("error getting tx blocks for double spend check spends", err)
	}
	mempoolTxRaws, err := item.GetMempoolTxRawByHashes(mempoolTxHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx blocks for double spend check spends", err)
	}
	var txRaws []*TxRaw
	for _, txBlockRaw := range txBlockRaws {
		txRaws = append(txRaws, &TxRaw{
			Hash: txBlockRaw.TxHash,
			Raw:  txBlockRaw.Raw,
		})
	}
	for _, mempoolTxRaw := range mempoolTxRaws {
		txRaws = append(txRaws, &TxRaw{
			Hash: mempoolTxRaw.TxHash,
			Raw:  mempoolTxRaw.Raw,
		})
	}
	return txRaws, nil
}
