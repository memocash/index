package tx_raw

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
)

type TxRaw struct {
	Hash []byte
	Raw  []byte
}

func Get(txHashes [][]byte) ([]*TxRaw, error) {
	txBlocks, err := chain.GetTxBlocks(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx blocks for double spend lock hashes", err)
	}
	var mempoolTxHashes [][]byte
Loop:
	for _, txHash := range txHashes {
		for _, txBlock := range txBlocks {
			if bytes.Equal(txBlock.TxHash[:], txHash) {
				continue Loop
			}
		}
		mempoolTxHashes = append(mempoolTxHashes, txHash)
	}
	var reqBlockTxs = make([]*item.ReqBlockTx, len(txBlocks))
	for i := range txBlocks {
		reqBlockTxs[i] = &item.ReqBlockTx{
			BlockHash: txBlocks[i].BlockHash[:],
			TxHash:    txBlocks[i].TxHash[:],
		}
	}
	txBlockRaws, err := item.GetRawTxBlocksByHashes(reqBlockTxs)
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
