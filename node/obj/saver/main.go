package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/dbi"
)

type Combined struct {
	Savers []dbi.TxSave
}

func (c *Combined) SaveTxs(block *wire.MsgBlock) error {
	for i, saver := range c.Savers {
		if err := saver.SaveTxs(block); err != nil {
			return jerr.Getf(err, "error saving transaction for saver %d", i)
		}
	}
	return nil
}

func NewCombined(savers []dbi.TxSave) *Combined {
	return &Combined{
		Savers: savers,
	}
}

func CombinedTxSaver(verbose bool) dbi.TxSave {
	return NewCombined([]dbi.TxSave{
		NewTxRaw(verbose),
		NewTx(verbose),
		NewUtxo(verbose),
		NewDoubleSpend(verbose),
	})
}
