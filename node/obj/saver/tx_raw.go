package saver

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"runtime"
	"time"
)

type TxRaw struct {
	Verbose bool
}

func (t *TxRaw) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	err := t.QueueTxs(block)
	if err != nil {
		return jerr.Get("error queueing msg txs", err)
	}
	return nil
}

func (t *TxRaw) QueueTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var blockHash chainhash.Hash
	var blockHashBytes []byte
	if !block.Header.Timestamp.IsZero() {
		blockHash = block.BlockHash()
		blockHashBytes = blockHash.CloneBytes()
	}
	if len(blockHashBytes) > 0 {
		jlog.Logf("block: %s, %s, txs: %10s, size: %14s\n", blockHash.String(),
			block.Header.Timestamp.Format("2006-01-02 15:04:05"), jfmt.AddCommasInt(len(block.Transactions)),
			jfmt.AddCommasInt(block.SerializeSize()))
	}
	seenTime := time.Now()
	var objects []item.Object
	var txsSize int
	for _, tx := range block.Transactions {
		raw := memo.GetRaw(tx)
		txHash := chainhash.DoubleHashH(raw)
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		if len(blockHashBytes) > 0 {
			objects = append(objects, &item.BlockTxRaw{
				BlockHash: blockHashBytes,
				TxHash:    txHashBytes,
				Raw:       raw,
			}, &item.BlockTx{
				TxHash:    txHashBytes,
				BlockHash: blockHashBytes,
			}, &item.TxBlock{
				TxHash:    txHashBytes,
				BlockHash: blockHashBytes,
			})
		} else {
			objects = append(objects, &item.MempoolTxRaw{
				TxHash: txHashBytes,
				Raw:    raw,
			})
		}
		objects = append(objects, &item.TxSeen{
			TxHash:    txHashBytes,
			Timestamp: seenTime,
		})
		txsSize += tx.SerializeSize()
		if len(objects) >= 10000 || txsSize > 10000000 {
			err := item.Save(objects)
			if err != nil {
				return jerr.Get("error saving db tx objects (at limit)", err)
			}
			objects = nil
			txsSize = 0
			runtime.GC()
		}
	}
	err := item.Save(objects)
	if err != nil {
		return jerr.Get("error saving db tx objects", err)
	}
	return nil
}

func NewTxRaw(verbose bool) *TxRaw {
	return &TxRaw{
		Verbose: verbose,
	}
}
