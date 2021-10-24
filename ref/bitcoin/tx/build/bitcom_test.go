package build_test

import (
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type BitcomTest struct {
	Request  build.BitcomRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (tst BitcomTest) Test(t *testing.T) {
	tx, err := build.Bitcom(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestBitcomSimple(t *testing.T) {
	BitcomTest{
		Request: build.BitcomRequest{
			Wallet:   test_tx.GetAddress1WalletSingle100k(),
			Filename: "Test.txt",
			Filetype: "plain/text",
			Contents: []byte("Test file"),
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "bfc0416ea628ffb819254815771d74159cc38f603ea3fed94b88175626eb35e5",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100e971fa48e781fe5783aba8ba82efc1df4cd54fd7b5abe887fabbf3d650da959102200847f38c6660c6d789ac37ba43a25134fd4671abe6ecb07c5d28e0e93ab46a41412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000436a2231394878696756345179427633744870515663554551797131707a5a56646f41757409546573742066696c650a706c61696e2f746578740008546573742e74787494850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
