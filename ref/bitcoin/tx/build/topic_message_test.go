package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type TopicMessageTest struct {
	Request  build.TopicMessageRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (test TopicMessageTest) Test(t *testing.T) {
	txs, err := build.TopicMessage(test.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    test.Error,
		TxHashes: test.TxHashes,
	}.Check(err, t)
}

func TestSimple(t *testing.T) {
	TopicMessageTest{
		Request: build.TopicMessageRequest{
			Wallet:    test_tx.GetAddress1WalletSingle100k(),
			TopicName: "test topic",
			Message:   "Topic message",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "53cc1932661c1d9afd34a5d3aa59a901fcb031cf903b6c2c298cd79c4d972954",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402200290f9c6668078231c5abd487f5ece5c03b9dc47a0e584bcfab6a2a0e033685c022077756417fa049e82c6768f0778c970da05c92998385d213321010ff18a546740412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000001d6a026d0c0a7465737420746f7069630d546f706963206d657373616765ba850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
