package saver

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jfmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/dbi"
	"runtime"
	"time"
)

type TxRaw struct {
	Verbose bool
}

func (t *TxRaw) SaveTxs(b *dbi.Block) error {
	if b.IsNil() {
		return jerr.Newf("error nil block")
	}
	if err := t.QueueTxs(b); err != nil {
		return jerr.Get("error queueing msg txs", err)
	}
	return nil
}

func (t *TxRaw) QueueTxs(block *dbi.Block) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var blockHash chainhash.Hash
	var blockHashBytes []byte
	if dbi.BlockHeaderSet(block.Header) {
		blockHash = block.Header.BlockHash()
		blockHashBytes = blockHash.CloneBytes()
	}
	if len(blockHashBytes) > 0 {
		jlog.Logf("block: %s, %s, txs: %10s\n", blockHash.String(),
			block.Header.Timestamp.Format("2006-01-02 15:04:05"), jfmt.AddCommasInt(len(block.Transactions)))
	}
	seenTime := time.Now()
	var objects []db.Object
	var txsSize int
	for i := range block.Transactions {
		raw := memo.GetRaw(block.Transactions[i].MsgTx)
		txHash := chainhash.DoubleHashH(raw)
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("raw tx: %s\n", txHash.String())
		}
		if len(blockHashBytes) > 0 {
			objects = append(objects, &item.BlockTxRaw{
				BlockHash: blockHashBytes,
				TxHash:    txHashBytes,
				Raw:       raw,
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
		txsSize += block.Transactions[i].MsgTx.SerializeSize()
		if len(objects) >= 25000 || txsSize > 250000000 {
			if err := db.Save(objects); err != nil {
				return jerr.Get("error saving db tx objects (at limit)", err)
			}
			objects = nil
			txsSize = 0
			runtime.GC()
		}
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db tx objects", err)
	}
	return nil
}

func NewTxRaw(verbose bool) *TxRaw {
	return &TxRaw{
		Verbose: verbose,
	}
}
