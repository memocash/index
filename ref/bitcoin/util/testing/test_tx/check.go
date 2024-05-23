package test_tx

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"log"
	"testing"
)

type TxHash struct {
	TxHash string
	TxRaw  string
}

type Checker struct {
	Name     string
	Txs      []*memo.Tx
	Error    error
	TxHashes []TxHash
}

func (c Checker) Check(err error, t *testing.T) {
	if err != nil {
		if errors.Is(err, c.Error) {
			if testing.Verbose() {
				log.Printf("Tx has expected error - %v\n", c.Error)
				log.Printf("%v\n", err)
			}
		} else {
			if c.Name != "" {
				t.Error(fmt.Errorf("error generating test tx(s) - Name: %s; %w", c.Name, err))
			} else {
				t.Error(fmt.Errorf("error generating test tx(s); %w", err))
			}
		}
		return
	}
	var isError bool
	if len(c.TxHashes) != len(c.Txs) {
		t.Error(fmt.Errorf("%s: len(c.TxHashes) [%d] != len(c.Txs) [%d]",
			c.Name, len(c.TxHashes), len(c.Txs)))
		isError = true
		goto done
	}
	for _, txHash := range c.TxHashes {
		raw, err := hex.DecodeString(txHash.TxRaw)
		if err != nil {
			t.Error(fmt.Errorf("error parsing raw tx; %w", err))
			isError = true
			goto done
		}
		msg, err := memo.GetMsgFromRaw(raw)
		if err != nil {
			t.Error(fmt.Errorf("error getting message from raw; %w", err))
			isError = true
			goto done
		}
		if txHash.TxHash != msg.TxHash().String() {
			t.Error(fmt.Errorf("error tx hash (%s) does not match msg tx hash (%s)",
				txHash.TxHash, msg.TxHash().String()))
			isError = true
			goto done
		}
	}
	for i, tx := range c.Txs {
		if tx.MsgTx.TxHash().String() != c.TxHashes[i].TxHash {
			t.Error(fmt.Errorf("%s: tx hash (%s) does not match expected (%s)",
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
