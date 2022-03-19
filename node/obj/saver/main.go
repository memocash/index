package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/dbi"
)

type CombinedTx struct {
	Savers []dbi.TxSave
}

func (c *CombinedTx) SaveTxs(block *wire.MsgBlock) error {
	for i, saver := range c.Savers {
		if err := saver.SaveTxs(block); err != nil {
			return jerr.Getf(err, "error saving transaction for saver %d", i)
		}
	}
	return nil
}

func NewCombined(savers []dbi.TxSave) *CombinedTx {
	return &CombinedTx{
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

func CombinedTxSaverBasic(verbose bool) dbi.TxSave {
	return NewCombined([]dbi.TxSave{
		NewTxRaw(verbose),
		NewTx(verbose),
	})
}

type CombinedBlock struct {
	Main   dbi.BlockSave
	Savers []dbi.BlockSave
}

func (c *CombinedBlock) SaveBlock(block wire.BlockHeader) error {
	for i, saver := range c.Savers {
		if err := saver.SaveBlock(block); err != nil {
			return jerr.Getf(err, "error saving block for saver %d", i)
		}
	}
	return nil
}

func (c *CombinedBlock) GetBlock(height int64) ([]byte, error) {
	block, err := c.Main.GetBlock(height)
	if err != nil {
		return nil, jerr.Get("error getting block for combined block saver", err)
	}
	return block, nil
}

func NewCombinedBlock(main dbi.BlockSave, savers []dbi.BlockSave) *CombinedBlock {
	return &CombinedBlock{
		Main:   main,
		Savers: savers,
	}
}

func CombinedBlockSaver(verbose bool) dbi.BlockSave {
	blockSaver := NewBlock(verbose)
	return NewCombinedBlock(blockSaver, []dbi.BlockSave{
		blockSaver,
		NewClearSuspect(verbose),
	})
}
