package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type ProfileTest struct {
	Request  build.ProfileRequest
	Error    error
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
			TxHash: "38ab87305d1148952600c8fe250154fab16e352950029d8e17c6967a00652d6a",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402204e7d597c1d7f7dee73997147a479e917c15a1b2b4fb00db5e5eef7ea4c7cd7eb02206536454410cf2c505f88b0f240e48875d60978e421668d616efb4ee5efd08099412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d050c50726f66696c652074657874c6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
