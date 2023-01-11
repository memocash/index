package get

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/node/act/tx_raw"
)

type Tx struct {
	TxHash    [32]byte
	BlockHash [32]byte
	Raw       []byte
}

func (t *Tx) Get() error {
	txBlocks, err := chain.GetSingleTxBlocks(t.TxHash)
	if err != nil {
		return jerr.Get("error getting tx blocks", err)
	}
	switch l := len(txBlocks); {
	case l == 1:
		t.BlockHash = txBlocks[0].BlockHash
	case l > 1:
		return jerr.Newf("error unexpected number of tx blocks returned: %d", len(txBlocks))
	}
	txRaw, err := tx_raw.GetSingle(t.TxHash)
	if err != nil {
		return jerr.Get("error getting tx raws for lock hashes double spend checks", err)
	}
	t.Raw = txRaw.Raw
	return nil
}

func NewTx(txHash [32]byte) *Tx {
	return &Tx{
		TxHash: txHash,
	}
}
