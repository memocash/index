package saver

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/op_return"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Memo struct {
	Verbose bool
	Shard   int
}

func (t *Memo) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var height int64
	if !block.Header.Timestamp.IsZero() {
		blockHash := block.BlockHash()
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
		for j := range tx.TxIn {
			address, err := wallet.GetAddressFromSignatureScript(tx.TxIn[j].SignatureScript)
			if err != nil {
				return jerr.Get("error getting address from signature script", err)
			}
			lockHash = script.GetLockHashForAddress(*address)
			if len(lockHash) > 0 {
				break
			}
		}
		if len(lockHash) == 0 {
			return jerr.New("error could not find input pk hash for memo")
		}
		for h := range tx.TxOut {
			for _, opReturnHandler := range opReturnHandlers {
				if !opReturnHandler.CanHandle(tx.TxOut[h].PkScript) {
					continue
				}
				pushData, err := txscript.PushedData(tx.TxOut[h].PkScript)
				if err != nil {
					return jerr.Get("error getting pushed data", err)
				}
				if err := opReturnHandler.Handle(op_return.Info{
					Height:   height,
					TxHash:   txHashBytes,
					LockHash: lockHash,
					PushData: pushData,
				}); err != nil {
					return jerr.Get("error handling op return", err)
				}
			}
		}
	}
	return nil
}

func NewMemoShard(verbose bool, shard int) *Memo {
	return &Memo{
		Verbose: verbose,
		Shard:   shard,
	}
}
