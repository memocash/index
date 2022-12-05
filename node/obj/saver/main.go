package saver

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/dbi"
	"reflect"
)

type CombinedTx struct {
	Savers []dbi.TxSave
}

func (c *CombinedTx) SaveTxs(block *dbi.Block) error {
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

func NewCombinedTx(verbose, initial bool) *CombinedTx {
	return NewCombined([]dbi.TxSave{
		NewTxMinimal(verbose),
		NewAddress(verbose, initial),
		NewMemo(verbose, initial),
	})
}
