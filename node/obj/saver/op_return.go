package saver

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/op_return"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/wallet"
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
		var addr *wallet.Addr
		var SetLockHash = func() error {
			if addr != nil {
				return nil
			}
			for j := range tx.TxIn {
				address, err := wallet.GetAddrFromUnlockScript(tx.TxIn[j].SignatureScript)
				if err != nil {
					//log.Printf("error getting address from unlock script; %v", err)
					continue
				}
				addr = address
			}
			return nil
		}
		for h := range tx.TxOut {
			for _, opReturnHandler := range opReturnHandlers {
				if !opReturnHandler.CanHandle(tx.TxOut[h].PkScript) {
					continue
				}
				if err := SetLockHash(); err != nil {
					return fmt.Errorf("error setting lock hash for op return tx; %w", err)
				}
				if addr == nil {
					if err := item.LogProcessError(&item.ProcessError{
						TxHash: txHash,
						Error:  fmt.Sprintf("error could not find input pk hash for op return: %s", txHash.String()),
					}); err != nil {
						return fmt.Errorf("error saving process error for op return without lock hash; %w", err)
					}
					break
				}
				pushData, err := txscript.PushedData(tx.TxOut[h].PkScript)
				if err != nil {
					return fmt.Errorf("error getting pushed data; %w", err)
				}
				if err := opReturnHandler.Handle(ctx, parse.OpReturn{
					Seen:     transaction.Seen,
					Saved:    transaction.Saved,
					TxHash:   txHash,
					Addr:     *addr,
					PushData: pushData,
					Outputs:  tx.TxOut,
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
