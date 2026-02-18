package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type AliasTest struct {
	Request  build.AliasRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst AliasTest) Test(t *testing.T) {
	tx, err := build.Alias(tst.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestAliasSimple(t *testing.T) {
	AliasTest{
		Request: build.AliasRequest{
			Wallet:  test_tx.GetAddress1WalletSingle100k(),
			Address: test_tx.Address1,
			Alias:   "Alias",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "a2afbbf96fd1bfb1564188de8afe3cb7847f98c4d5140c2359de20edb97f4137",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100f8719e414323179ecaa865c23ff301999d39808ecd6209f52d2a297a07b90e79022021544e6db61832c15de055812da22eefc6b3861ef832a38ca2bfcf37f2e37e73412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000001f6a026d2614fc393e225549da044ed2c0011fd6c8a799806b6205416c696173b8850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
