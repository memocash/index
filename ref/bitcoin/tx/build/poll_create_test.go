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
			TxHash: "5bdb592378edddb0c262af2913e94a18a068b4523dcb54058aa6768110857c61",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402203a77cb92bf42595b64235a195169d2b4d9a5e9e18d173d967abc198cb44c2aaf0220675d39d8721aca5f703527eeaccd92b9491d0206a8e89d96877354cb8bd4ea2a412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d1051520a5465737420706f6c6c3fc6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
