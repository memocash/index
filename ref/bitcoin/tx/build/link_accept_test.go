package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type LinkAcceptTest struct {
	Request  build.LinkAcceptRequest
	Error    error
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
			TxHash: "9914705759e730e999775d5ae7e8b70036e0d30a0af8e43012626a1051125ade",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022039bcaacc30119be82df4d9abca1ebdfaceafe558ab1ccaf241a2ce9e76d137840220134ca50750670887a2893ee57345d761a83e252bb9a0ca4115a3b06b127f3dde412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d2120b158efa8e85ef8283481e000f9fb13b12599a8fa58fce12633093762ebd1cb75b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
