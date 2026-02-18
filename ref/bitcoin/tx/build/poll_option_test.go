package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PollOptionTest struct {
	Request  build.PollOptionRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst PollOptionTest) Test(t *testing.T) {
	tx, err := build.PollOption(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestPollOptionSimple(t *testing.T) {
	PollOptionTest{
		Request: build.PollOptionRequest{
			Wallet:     test_tx.GetAddress1WalletSingle100k(),
			PollTxHash: test_tx.GenericTxHash0,
			Option:     "Option 1",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "7ecde435509e45bbba0a86ae1d77738aa6f41388873c67da208b2ff0c54ce405",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100ee1107c4abb4ac70bed5d03ba34b2259f46fd4fbeaf185d5693e32a04335a5f802204b4c0d6b8242919360ad011cf1e1c81f3ce57ad38664855d21ac9374910f98fb412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000002e6a026d132075cbd1eb6237093326e1fc58faa89925b113fbf900e0813428f85ee8a8ef58b1084f7074696f6e2031a9850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
