package tx_raw

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
)

type TxRaw struct {
	Hash [32]byte
	Raw  []byte
}

func Get(txHashes [][32]byte) ([]*TxRaw, error) {
	txBlocks, err := chain.GetTxBlocks(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx blocks for double spend lock hashes", err)
	}
	var mempoolTxHashes [][32]byte
Loop:
	for _, txHash := range txHashes {
		for _, txBlock := range txBlocks {
			if txBlock.TxHash == txHash {
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
	mempoolTxRaws, err := item.GetMempoolTxRawByHashes(db.FixedTxHashesToRaw(mempoolTxHashes))
	if err != nil {
		return nil, jerr.Get("error getting tx blocks for double spend check spends", err)
	}
	var txRaws []*TxRaw
	addTxRaw := func(txHash []byte, raw []byte) {
		txRaw := &TxRaw{Raw: raw}
		copy(txRaw.Hash[:], txHash)
		txRaws = append(txRaws, txRaw)
	}
	for _, txBlockRaw := range txBlockRaws {
		addTxRaw(txBlockRaw.TxHash, txBlockRaw.Raw)
	}
	for _, mempoolTxRaw := range mempoolTxRaws {
		addTxRaw(mempoolTxRaw.TxHash, mempoolTxRaw.Raw)
	}
	return txRaws, nil
}
