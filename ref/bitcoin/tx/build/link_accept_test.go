package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type LinkAcceptTest struct {
	Request  build.LinkAcceptRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (tst LinkAcceptTest) Test(t *testing.T) {
	tx, err := build.LinkAccept(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestLinkAcceptSimple(t *testing.T) {
	LinkAcceptTest{
		Request: build.LinkAcceptRequest{
			Wallet:        test_tx.GetAddress1WalletSingle100k(),
			RequestTxHash: test_tx.GenericTxHash0,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "c1e8b1a1c0383ef5341b52751a0be217c466359b51fda9d08b07c17889df6aa4",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100865fafe25bffc6f5b6c8698255e10c33810f7050a9bfcf5573a28009ddba3ef302200e98003415411933439c3e8a639b2c264aaab5dc8f2e1f4f953d4dddf8f8e294412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d2120b158efa8e85ef8283481e000f9fb13b12599a8fa58fce12633093762ebd1cb75b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
