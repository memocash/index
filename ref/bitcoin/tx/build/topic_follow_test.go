package build_test

import (
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type TopicFollowTest struct {
	Request  build.TopicFollowRequest
	Error    string
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
			TxHash: "547d326425b5238eca20e3b8a792deda378e2a4068b4f2c3b395eea30c396d68",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b4830450221008fb82123a8386a2973b73f7f38798b0d3c3abd5233669dead82085ec70ba2ba602207f08bbf6fbb6360fe95dffde6f2998de25ca29d9e2da2c3713c8e472020b4f26412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000000f6a026d0d0a5465737420546f706963c8850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "f31bed2bd7de9c0e798810d9e27a06f62eb6567c02e69a8f7debdc7d0e3d23f8",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022009f1f6ce044cbbcbdbf8df2c484f93d3890a955103c4725aec7b64a5761d7990022047d18708415288c43464b0b2fb767e6af53acc9291cdd888975c53a463e70a82412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000000f6a026d0e0a5465737420546f706963c8850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
