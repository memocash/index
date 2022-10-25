package saver

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/obj/op_return"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
)

type Memo struct {
	Verbose bool
}

func (t *Memo) SaveTxs(block *dbi.Block) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var height int64
	if !block.Header.Timestamp.IsZero() {
		blockHash := block.Header.BlockHash()
		blockHashBytes := blockHash.CloneBytes()
		blockHeight, err := item.GetBlockHeight(blockHashBytes)
		if err != nil {
			return jerr.Get("error getting block height for memo", err)
		}
		height = blockHeight.Height
	}
	if height == 0 {
		height = item.HeightMempool
	}
	opReturnHandlers, err := op_return.GetHandlers()
	if err != nil {
		return jerr.Get("error getting op returns", err)
	}
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		var lockHash []byte
		var SetLockHash = func() error {
			if lockHash != nil {
				return nil
			}
			for j := range tx.TxIn {
				address, err := wallet.GetAddressFromSignatureScript(tx.TxIn[j].SignatureScript)
				if err != nil {
					return jerr.Get("error getting address from signature script", err)
				}
				lockHash = script.GetLockHashForAddress(*address)
				if len(lockHash) > 0 {
					// TODO: This is only needed temporarily because lock addresses are saved in UTXO processor
					//    	 which isn't being restarted. Remove eventually.
					var lockAddress = &item.LockAddress{
						LockHash: lockHash,
						Address:  address.GetEncoded(),
					}
					if err := db.Save([]db.Object{lockAddress}); err != nil {
						return jerr.Get("error saving db lock address object for op return tx", err)
					}
					break
				}
			}
			return nil
		}
		for h := range tx.TxOut {
			for _, opReturnHandler := range opReturnHandlers {
				if !opReturnHandler.CanHandle(tx.TxOut[h].PkScript) {
					continue
				}
				if err := SetLockHash(); err != nil {
					return jerr.Get("error setting lock hash for op return tx", err)
				}
				if len(lockHash) == 0 {
					if err := item.LogProcessError(&item.ProcessError{
						TxHash: txHashBytes,
						Error:  fmt.Sprintf("error could not find input pk hash for memo: %s", txHash.String()),
					}); err != nil {
						return jerr.Get("error saving process error for op return without lock hash", err)
					}
					break
				}
				pushData, err := txscript.PushedData(tx.TxOut[h].PkScript)
				if err != nil {
					return jerr.Get("error getting pushed data", err)
				}
				if err := opReturnHandler.Handle(parse.OpReturn{
					Height:   height,
					TxHash:   txHashBytes,
					LockHash: lockHash,
					PushData: pushData,
					Outputs:  tx.TxOut,
				}); err != nil {
					return jerr.Get("error handling op return", err)
				}
			}
		}
	}
	return nil
}

func NewMemo(verbose bool) *Memo {
	return &Memo{
		Verbose: verbose,
	}
}
