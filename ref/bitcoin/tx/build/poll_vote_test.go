package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PollVoteTest struct {
	Request  build.PollVoteRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (tst PollVoteTest) Test(t *testing.T) {
	txs, err := build.PollVote(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestPollVoteSimple(t *testing.T) {
	PollVoteTest{
		Request: build.PollVoteRequest{
			Wallet:           test_tx.GetAddress1WalletSingle100k(),
			PollOptionTxHash: test_tx.HashEmptyTx,
			Message:          "test",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "7fe0a0e2ca0eaa56edcba71431bf3e6612fcf7e0e98b8138da25f8c3927e420c",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100b7547363030d87b578cd39ba6875fa005491512b9784defbe66db0903edfc37b0220391d55292efcc29b9dfcd5af1889155772bd67c5f519142a213f5e40f73b5226412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000002a6a026d142043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d20474657374ad850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestPollVoteWithTip(t *testing.T) {
	PollVoteTest{
		Request: build.PollVoteRequest{
			Wallet:           test_tx.GetAddress1WalletSingle100k(),
			PollOptionTxHash: test_tx.HashEmptyTx,
			Message:          "test",
			Tip:              10000,
			TipAddress:       test_tx.Address2,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "9d5195d797cc02015f8d7ad4ca1a68e5a8b643315e720930c6223f40018380f4",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402200f6bad89dd2089ab88ff3f4ca63ebea8e250000e0d5623d62dc99c1319f6fc5402207e78992b2a97c7b864cc8041ac2fdcbed521c25250a7a73ab117eab182d7cba2412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0300000000000000002a6a026d142043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2047465737410270000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac7b5e0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
