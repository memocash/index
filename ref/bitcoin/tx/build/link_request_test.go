package build_test

import (
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type LinkRequestTest struct {
	Request  build.LinkRequestRequest
	Error    string
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
			TxHash: "bd50a4617233852feb034c20289a949112fe8e79a3440888c70b7f3c44c96320",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022058e0b0d1a2df56ea711be9964801202aa534de0bc2dfc95d7a4af3bd5c68895602204ebbe620cce8fc1c6f103dcdd899c95eea2f30ff4c21a4ad0064c5111fd46db9412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000226a026d20140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105084d656d6f2d4f6c64b5850100000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
		}},
	}.Test(t)
}
