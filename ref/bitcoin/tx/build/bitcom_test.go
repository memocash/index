package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type BitcomTest struct {
	Request  build.BitcomRequest
	Error    error
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
			TxHash: "e654239d83d9fef25f2b4423696f1284c06df04611c7e44db3c204fa8c27511e",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022026434c2dbb94a263134125b9acb157d604e43b24013be1f511e8b73c7886b77202201b87d345b88cde22c7010d81e565c82368dc3134a2b5b8d44c2e191e142e4c3a412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000436a2231394878696756345179427633744870515663554551797131707a5a56646f41757409546573742066696c650a706c61696e2f746578740008546573742e74787494850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
