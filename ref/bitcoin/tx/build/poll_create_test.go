package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PollCreateTest struct {
	Request  build.PollCreateRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst PollCreateTest) Test(t *testing.T) {
	tx, err := build.PollCreate(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestPollCreateSimple(t *testing.T) {
	PollCreateTest{
		Request: build.PollCreateRequest{
			Wallet:      test_tx.GetAddress1WalletSingle100k(),
			PollType:    memo.PollTypeOne,
			Question:    "Test poll?",
			OptionCount: 2,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "631c34ca853eafaf160d905f4102c0fb3e7f99e2f8f6f6042055248fb07a5eaf",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402200a822a2b721f31062a98e5f630b56155a44910c1e6a831c8422963d2cbc5f691022026a80b4db31d5aba48b8ca5cbb5393ab5bea9ba4fdd74d34aa38cb580cc049f6412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d1051520a5465737420706f6c6c3fc6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
