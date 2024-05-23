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
			TxHash: "3e9b0fba675ef33033ed69beb4f2858bb7e5461b0be39d78ecafab9b54a45543",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100a59e9ae319e43d462c00f2053d692343dffe6c2adfea6244c77d9dcf9816b36a0220228520b5c0959e7ef5abbe76b97194bf37c661089c7615a44390ef82a328fe9a412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d06140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "d03513888712619f89a912d6a3c48fff9c69dd29893a86d74f79a37321fb8510",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100e7f3db5f37aa8e984ef7764f9e0479eaff1593522c5fdd917ea52c916366b0de0220756dcbb6cf05bd25c496398d3f35db31037c3074b5b42c8b6af5b619bd2922b8412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d07140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
