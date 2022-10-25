package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/dbi"
	"reflect"
)

type CombinedTx struct {
	Savers []dbi.TxSave
}

func (c *CombinedTx) SaveTxs(block *wire.MsgBlock) error {
	for _, saver := range c.Savers {
		if err := saver.SaveTxs(block); err != nil {
			return jerr.Getf(err, "error saving transaction for saver - %s", reflect.TypeOf(saver))
		}
	}
	return nil
}

func NewCombined(savers []dbi.TxSave) *CombinedTx {
	return &CombinedTx{
		Savers: savers,
	}
}

func NewCombinedBlockTxRaw(verbose bool) *CombinedTx {
	return NewCombined([]dbi.TxSave{
		NewTxRaw(verbose),
	})
}

func NewCombinedAll(verbose bool) *CombinedTx {
	return NewCombined([]dbi.TxSave{
		NewTxRaw(verbose),
		NewTx(verbose),
		NewUtxo(verbose),
		NewLockHeight(verbose),
		NewDoubleSpend(verbose),
		NewMemo(verbose),
	})
}

func NewCombinedTx(verbose bool) *CombinedTx {
	return NewCombined([]dbi.TxSave{
		NewTxRaw(verbose),
		NewTx(verbose),
	})
}

func NewCombinedOutput(verbose bool) *CombinedTx {
	return NewCombined([]dbi.TxSave{
		NewLockHeight(verbose),
		NewDoubleSpend(verbose),
		NewMemo(verbose),
	})
}
