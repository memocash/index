package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type LinkRevokeTest struct {
	Request  build.LinkRevokeRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst LinkRevokeTest) Test(t *testing.T) {
	tx, err := build.LinkRevoke(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestLinkRevokeSimple(t *testing.T) {
	LinkRevokeTest{
		Request: build.LinkRevokeRequest{
			Wallet:       test_tx.GetAddress1WalletSingle100k(),
			AcceptTxHash: test_tx.GenericTxHash0,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "b9508521ba1d7aabf3ea2077c7fec69cf33a7aecd4d477f85a4146cb629e5c6f",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100bb8c5e278b3fcc3a6e2ac03278d03865ec3558a19d6b6045b9701cd22d41084202206b429c525354d650fdef2a4e8350c99702824455b62489d6364d4fc17fe50b19412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d2220b158efa8e85ef8283481e000f9fb13b12599a8fa58fce12633093762ebd1cb75b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
