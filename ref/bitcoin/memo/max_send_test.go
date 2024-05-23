package memo_test

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"log"
	"testing"
)

type MaxSendTest struct {
	UTXOs   []memo.UTXO
	MaxSend int64
}

func (tst MaxSendTest) Test(t *testing.T) {
	maxSend := memo.GetMaxSendForUTXOs(tst.UTXOs)
	if maxSend != tst.MaxSend {
		t.Error(fmt.Errorf("MaxSend %d does not match expected %d", maxSend, tst.MaxSend))
	}
	if testing.Verbose() {
		log.Printf("MaxSend %d, expected %d\n", maxSend, tst.MaxSend)
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
