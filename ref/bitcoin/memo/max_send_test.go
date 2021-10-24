package memo_test

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type MaxSendTest struct {
	UTXOs   []memo.UTXO
	MaxSend int64
}

func (tst MaxSendTest) Test(t *testing.T) {
	maxSend := memo.GetMaxSendForUTXOs(tst.UTXOs)
	if maxSend != tst.MaxSend {
		t.Error(jerr.Newf("MaxSend %d does not match expected %d", maxSend, tst.MaxSend))
	}
	if testing.Verbose() {
		jlog.Logf("MaxSend %d, expected %d\n", maxSend, tst.MaxSend)
	}
}

func TestMaxSendSimple(t *testing.T) {
	MaxSendTest{
		UTXOs:   test_tx.GetUtxosTestSet1(),
		MaxSend: test_tx.UtxosTestSet1MaxSend,
	}.Test(t)
}

func TestMaxSendDust(t *testing.T) {
	MaxSendTest{
		UTXOs:   []memo.UTXO{test_tx.GetUTXO(memo.DustMinimumOutput)},
		MaxSend: 0,
	}.Test(t)
}
