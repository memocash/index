package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PollVoteTest struct {
	Request  build.PollVoteRequest
	Error    error
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
			TxHash: "740d4d9c199f362c011a53fa186b93d4068eb93664a1f089685f6c74d1fd3b90",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220223947a10b4b646bcd2ff97b0495e41873d7421e04f1f19e0e48011c161ea8a802203c37607895196c26af21695da7940564357953b6f30d64422ebd0d1fb9eda993412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000002a6a026d142043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d20474657374ad850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "7db3949b059bf99bbc22800a4c4cb4e05b58370835d56e0f5f4772498bbfce7b",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100dcd6c8967d4f518adc2c94e89b9e93d5c9bc26a56aa31ecdbb8768131a662c1202205062054e02ef1d43fc21dc385a4f50990761961d786df140d53744b17665229b412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0300000000000000002a6a026d142043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2047465737410270000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac7b5e0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
