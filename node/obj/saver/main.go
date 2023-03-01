package saver

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/dbi"
	"reflect"
	"time"
)

type CombinedTx struct {
	Savers    []dbi.TxSave
	SaveTimes map[string]time.Duration
}

func (c *CombinedTx) SaveTxs(block *dbi.Block) error {
	for _, saver := range c.Savers {
		start := time.Now()
		if err := saver.SaveTxs(block); err != nil {
			return jerr.Getf(err, "error saving transaction for saver - %s", reflect.TypeOf(saver))
		}
		c.SaveTimes[reflect.TypeOf(saver).String()] = time.Since(start)
	}
	return nil
}

func NewCombined(savers []dbi.TxSave) *CombinedTx {
	return &CombinedTx{
		Savers:    savers,
		SaveTimes: make(map[string]time.Duration),
	}
}

func NewCombinedTx(verbose bool) *CombinedTx {
	return NewCombined([]dbi.TxSave{
		NewTxMinimal(verbose),
		NewAddress(verbose),
		NewOpReturn(verbose),
		NewTxProcessed(verbose),
	})
}
