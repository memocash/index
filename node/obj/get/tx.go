package get

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

type Tx struct {
	TxHash    []byte
	BlockHash []byte
	Raw       []byte
}

func (t *Tx) Get() error {
	txBlocks, err := item.GetSingleTxBlocks(t.TxHash)
	if err != nil {
		return jerr.Get("error getting tx blocks", err)
	}
	switch l := len(txBlocks); {
	case l == 0:
		mempoolTx, err := item.GetMempoolTxRawByHash(t.TxHash)
		if err != nil {
			return jerr.Get("error getting mempool tx raw and tx block not found", err)
		}
		t.Raw = mempoolTx.Raw
		return nil
	case l != 1:
		return jerr.Newf("error unexpected number of tx blocks returned: %d", len(txBlocks))
	}
	txRaw, err := item.GetRawBlockTxByHash(txBlocks[0].BlockHash, txBlocks[0].TxHash)
	if err != nil {
		return jerr.Getf(err, "error getting raw block tx by hash: %s %s",
			hs.GetTxString(txBlocks[0].BlockHash), hs.GetTxString(txBlocks[0].TxHash))
	}
	t.BlockHash = txBlocks[0].BlockHash
	t.Raw = txRaw.Raw
	return nil
}

func NewTx(txHash []byte) *Tx {
	return &Tx{
		TxHash: txHash,
	}
}
