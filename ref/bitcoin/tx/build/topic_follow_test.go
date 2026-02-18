package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type TopicFollowTest struct {
	Request  build.TopicFollowRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst TopicFollowTest) Test(t *testing.T) {
	txs, err := build.TopicFollow(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestTopicFollowSimple(t *testing.T) {
	TopicFollowTest{
		Request: build.TopicFollowRequest{
			Wallet:    test_tx.GetAddress1WalletSingle100k(),
			TopicName: "Test Topic",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "f2f2e03beb69c0da33f1d3581a0f374441c007bfe8d7d1c3d8b2ce66aaa562a1",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022027a79d86e9acf9aeb2e80ef9ce321be2cfc0d331a64c801f69e22c4378ed4cca0220579b38f031412ebcf0d2b41ccfa332d0adfb63845caf0bf34e7486e6e29c5f49412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000000f6a026d0d0a5465737420546f706963c8850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestTopicUnfollowSimple(t *testing.T) {
	TopicFollowTest{
		Request: build.TopicFollowRequest{
			Wallet:    test_tx.GetAddress1WalletSingle100k(),
			TopicName: "Test Topic",
			Unfollow:  true,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "99a56d4c97d9e23fe25fa7afc24a0792c69853093b2d7f252c863442bd04b962",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100bc4e5d4d551d6a1769eca8f06a02d553e34df5978727fd00ac3501c9cc02d98802200ad688a7e6ae95998502f4aae16b47b2dba07a965eb2e86f23e478ce22ada87a412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000000f6a026d0e0a5465737420546f706963c8850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
