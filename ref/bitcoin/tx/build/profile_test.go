package build_test

import (
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type ProfileTest struct {
	Request  build.ProfileRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (tst ProfileTest) Test(t *testing.T) {
	txs, err := build.Profile(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestProfileSimple(t *testing.T) {
	ProfileTest{
		Request: build.ProfileRequest{
			Wallet: test_tx.GetAddress1WalletSingle100k(),
			Text:   "Profile text",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "6cdf4a98fdd33b6b27c40183b93de9c594c6e405aded6f7db2270814bd0b3ba9",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402206afaf8597de1678b098b70bff2504ed5e544d3d4ad11c68c0aac8ef2006dd6c702206c099792739a9191ea490aadbe655b8743993fc39d3f809215e32e3e7582b1a8412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d050c50726f66696c652074657874c6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
