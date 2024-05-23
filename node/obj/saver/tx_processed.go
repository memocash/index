package saver

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/dbi"
	"log"
	"time"
)

type TxProcessed struct {
	Verbose bool
}

func (t *TxProcessed) SaveTxs(ctx context.Context, block *dbi.Block) error {
	if block.IsNil() {
		return fmt.Errorf("error nil block")
	}
	if err := t.QueueTxs(block); err != nil {
		return fmt.Errorf("error queueing tx minimal block; %w", err)
	}
	return nil
}

func (t *TxProcessed) QueueTxs(block *dbi.Block) error {
	processedTime := time.Now()
	var objects []db.Object
	for _, dbiTx := range block.Transactions {
		txHash := chainhash.Hash(dbiTx.Hash)
		if t.Verbose {
			log.Printf("processed tx: %s\n", txHash.String())
		}
		objects = append(objects, &chain.TxProcessed{
			TxHash:    txHash[:],
			Timestamp: processedTime,
		})
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving db tx objects; %w", err)
	}
	return nil
}

func NewTxProcessed(verbose bool) *TxProcessed {
	return &TxProcessed{
		Verbose: verbose,
	}
}
