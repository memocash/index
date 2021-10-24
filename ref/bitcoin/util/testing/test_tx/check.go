package test_tx

import (
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/parse"
	"testing"
)

type TxHash struct {
	TxHash string
	TxRaw  string
}

type Checker struct {
	Name     string
	Txs      []*memo.Tx
	Error    string
	TxHashes []TxHash
}

func (c Checker) Check(err error, t *testing.T) {
	if err != nil {
		if c.Error != "" && jerr.HasError(err, c.Error) {
			if testing.Verbose() {
				jlog.Logf("Tx has expected error - '%s'\n", c.Error)
				jerr.Create(err).Print()
			}
		} else {
			if c.Name != "" {
				t.Error(jerr.Getf(err, "error generating test tx(s) - Name: %s", c.Name))
			} else {
				t.Error(jerr.Get("error generating test tx(s)", err))
			}
		}
		return
	}
	var isError bool
	if len(c.TxHashes) != len(c.Txs) {
		t.Error(jerr.Newf("%s: len(c.TxHashes) [%d] != len(c.Txs) [%d]",
			c.Name, len(c.TxHashes), len(c.Txs)))
		isError = true
		goto done
	}
	for _, txHash := range c.TxHashes {
		raw, err := hex.DecodeString(txHash.TxRaw)
		if err != nil {
			t.Error(jerr.Get("error parsing raw tx", err))
			isError = true
			goto done
		}
		msg, err := memo.GetMsgFromRaw(raw)
		if err != nil {
			t.Error(jerr.Get("error getting message from raw", err))
			isError = true
			goto done
		}
		if txHash.TxHash != msg.TxHash().String() {
			t.Error(jerr.Newf("error tx hash (%s) does not match msg tx hash (%s)",
				txHash.TxHash, msg.TxHash().String()))
			isError = true
			goto done
		}
	}
	for i, tx := range c.Txs {
		if tx.MsgTx.TxHash().String() != c.TxHashes[i].TxHash {
			t.Error(jerr.Newf("%s: tx hash (%s) does not match expected (%s)",
				c.Name, tx.MsgTx.TxHash().String(), c.TxHashes[i].TxHash))
			isError = true
			goto done
		}
	}
done:
	if testing.Verbose() || isError {
		for _, tx := range c.Txs {
			txInfo := parse.GetTxInfo(tx)
			txInfo.Print()
		}
	}
}
