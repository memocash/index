package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PollOptionTest struct {
	Request  build.PollOptionRequest
	Error    string
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
			TxHash: "f00b6df10c9ab1117c3254e9cf8e1a74a5e36798ffb69180b3c2a87a99255385",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402204a4bb0213d4026c4028b07359f8bfd26c166f9153eb2754d98576da1ee154aa30220031d623af2c7453030c10e0510cb465ec0ca6eb94a449cd272e666fb1dd2f6c6412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000002e6a026d132075cbd1eb6237093326e1fc58faa89925b113fbf900e0813428f85ee8a8ef58b1084f7074696f6e2031a9850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
