package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type LinkRequestTest struct {
	Request  build.LinkRequestRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst LinkRequestTest) Test(t *testing.T) {
	tx, err := build.LinkRequest(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestLinkRequestSimple(t *testing.T) {
	LinkRequestTest{
		Request: build.LinkRequestRequest{
			OldWallet: test_tx.GetAddress1WalletSingle100k(),
			NewWallet: test_tx.GetAddress2WalletEmpty(),
			Message:   "Memo-Old",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "5a2caa17911414b301f64233a657511b6cad095b32404df4c90b38aed19ff330",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100ed75fe4ec0b1b19c851dafa9acdb158e954c47c3b617847dc112c0f1a8113e5202207f185817c468699e4c7c02491ccdec3410365186c5afa1771e54835e2a13a8b3412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000226a026d20140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105084d656d6f2d4f6c64b5850100000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
		}},
	}.Test(t)
}
