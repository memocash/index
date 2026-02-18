package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type SetNameTest struct {
	Request  build.SetNameRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst SetNameTest) Test(t *testing.T) {
	txs, err := build.SetName(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestSetNameSimple(t *testing.T) {
	SetNameTest{
		Request: build.SetNameRequest{
			Wallet: test_tx.GetAddress1WalletSingle100k(),
			Name:   "SetName name",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "cd730ff95f2b3180f4e01762854a4403072dde43a71bfacb095090f310d410e8",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100bc87195bd855f6c34c3a8aefc4695f6c8fad7f817673a8e219d80e4e865c0d4302203f50ff5c81adc52add90a92bc3d6acd52b432ea10cf453f872ebef1fb4bf9a69412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d010c5365744e616d65206e616d65c6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
