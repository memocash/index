package saver

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/node/obj/op_return"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/dbi"
	"log"
)

type OpReturn struct {
	Verbose bool
}

func (r *OpReturn) SaveTxs(ctx context.Context, b *dbi.Block) error {
	if b.IsNil() {
		return fmt.Errorf("error nil block")
	}
	opReturnHandlers, err := op_return.GetHandlers()
	if err != nil {
		return fmt.Errorf("error getting op returns; %w", err)
	}
	for _, transaction := range b.Transactions {
		var tx = transaction.MsgTx
		txHash := tx.TxHash()
		if r.Verbose {
			log.Printf("tx: %s\n", txHash.String())
		}
		for h := range tx.TxOut {
			for _, opReturnHandler := range opReturnHandlers {
				if !opReturnHandler.CanHandle(tx.TxOut[h].PkScript) {
					continue
				}
				pushData, err := txscript.PushedData(tx.TxOut[h].PkScript)
				if err != nil {
					return fmt.Errorf("error getting pushed data; %w", err)
				}
				if err := opReturnHandler.Handle(ctx, parse.OpReturn{
					Seen:        transaction.Seen,
					Saved:       transaction.Saved,
					TxHash:      txHash,
					PushData:    pushData,
					Outputs:     tx.TxOut,
					Inputs:      tx.TxIn,
					OutputIndex: h,
				}); err != nil {
					return fmt.Errorf("error handling op return; %w", err)
				}
			}
		}
	}
	return nil
}

func NewOpReturn(verbose bool) *OpReturn {
	return &OpReturn{
		Verbose: verbose,
	}
}
