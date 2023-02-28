package saver

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/dbi"
	"time"
)

type TxProcessed struct {
	Verbose bool
}

func (t *TxProcessed) SaveTxs(block *dbi.Block) error {
	if block.IsNil() {
		return jerr.Newf("error nil block")
	}
	if err := t.QueueTxs(block); err != nil {
		return jerr.Get("error queueing tx minimal block", err)
	}
	return nil
}

func (t *TxProcessed) QueueTxs(block *dbi.Block) error {
	processedTime := time.Now()
	var objects []db.Object
	for _, dbiTx := range block.Transactions {
		txHash := chainhash.Hash(dbiTx.Hash)
		if t.Verbose {
			jlog.Logf("processed tx: %s\n", txHash.String())
		}
		objects = append(objects, &chain.TxProcessed{
			TxHash:    txHash[:],
			Timestamp: processedTime,
		})
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db tx objects", err)
	}
	return nil
}

func NewTxProcessed(verbose bool) *TxProcessed {
	return &TxProcessed{
		Verbose: verbose,
	}
}
