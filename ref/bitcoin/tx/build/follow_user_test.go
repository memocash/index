package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type FollowTest struct {
	Request  build.FollowUserRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst FollowTest) Test(t *testing.T) {
	txs, err := build.FollowUser(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestFollowSimple(t *testing.T) {
	FollowTest{
		Request: build.FollowUserRequest{
			Wallet:     test_tx.GetAddress1WalletSingle100k(),
			UserPkHash: test_tx.Address2pkHash,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "59ee1dc674ce665862e5f4f9dcfcfb96aa01338cbedb47088ddaf43f8239109d",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100f8cb2bef0527484b344e966b0f3b120ab15c885521765832f8b8c2965dc14d140220692228a54a09ce98f678e4da46c5c2c33004eacaee6efb5db68e0416345f8086412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d06140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestUnfollowSimple(t *testing.T) {
	FollowTest{
		Request: build.FollowUserRequest{
			Wallet:     test_tx.GetAddress1WalletSingle100k(),
			UserPkHash: test_tx.Address2pkHash,
			Unfollow:   true,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "b30b0c0adcdfa5618ece5f074d765370382c48f05361c3a52c9f09608eb49785",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b48304502210090c7e9d47465d327607578c52eb00aef5d5bb16a97027478bfe91ae4116312e9022022f786fffd954efdf45c259a8d8c3fb5493ec5d84d25358471749847ddbab45b412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d07140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
