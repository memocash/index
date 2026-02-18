package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PostTest struct {
	Request  build.PostRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst PostTest) Test(t *testing.T) {
	txs, err := build.Post(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestPostSimple(t *testing.T) {
	PostTest{
		Request: build.PostRequest{
			Wallet:  test_tx.GetAddress1WalletSingle100k(),
			Message: "Post message",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "e83d7bfdf6529d8484d4b54bcd9f7c4b4b0923409458005d6df69442f661a2e7",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022050052efe87d5bf493065909c173c4c0485e247f665d5ceb15b9507ef80ee76f8022056c4116d424a18e9cf0c8ffde9dfa0ae8ac9cb34e82bcb9b0bbc0683896fc932412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d020c506f7374206d657373616765c6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
