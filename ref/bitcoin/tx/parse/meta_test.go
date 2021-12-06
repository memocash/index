package parse_test

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type MetaTest struct {
	Tx   *wire.MsgTx
	Type memo.OutputType
}

func (tst MetaTest) Test(t *testing.T) {
	meta := parse.GetMeta(tst.Tx)
	if meta.Multi {
		t.Error(jerr.New("error meta found multi"))
	}
	jlog.Logf("meta: %t\n", meta.OpReturn != nil)
	if meta.OutputType != tst.Type {
		t.Error(jerr.Newf("meta.OutputType %s does not match expected %s", meta.OutputType, tst.Type))
	}
	if testing.Verbose() {
		jlog.Logf("meta.OutputType %s, expected %s\n", meta.OutputType, tst.Type)
	}
}

func TestMetaSimple(t *testing.T) {
	scr, _ := script.Post{Message: test_tx.TestMessage}.Get()
	tx := &wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: scr,
		}},
	}
	MetaTest{
		Tx:   tx,
		Type: memo.OutputTypeMemoMessage,
	}.Test(t)
}

func TestMetaSlpSend(t *testing.T) {
	scr, _ := script.TokenSend{
		TokenHash:  test_tx.GenericTxHash0,
		SlpType:    memo.SlpDefaultTokenType,
		Quantities: []uint64{10000},
	}.Get()
	tx := &wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: scr,
		}},
	}
	MetaTest{
		Tx:   tx,
		Type: memo.OutputTypeTokenSend,
	}.Test(t)
}

func TestMetaSlpSendCreate(t *testing.T) {
	scr, _ := script.TokenCreate{
		Ticker:   test_tx.TestTokenTicker,
		Name:     test_tx.TestTokenName,
		SlpType:  memo.SlpDefaultTokenType,
		Quantity: 10000,
	}.Get()
	tx := &wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: scr,
		}},
	}
	MetaTest{
		Tx:   tx,
		Type: memo.OutputTypeTokenCreate,
	}.Test(t)
}
