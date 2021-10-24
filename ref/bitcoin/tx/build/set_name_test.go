package build_test

import (
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type SetNameTest struct {
	Request  build.SetNameRequest
	Error    string
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
			TxHash: "2b755d0cf2c8a63c2779553028dcd9f84746c9bf7048f42cc5a11f0b9ea96592",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100bc3fc1ce800747dfedd4ce6425b5ecf44d4f3d736a4c1f48750b84bf8583d8fd02207c121ed06ca5e888b6dd41e0558c9b6350b65b7d2211d438edf9c592390ee26e412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d010c5365744e616d65206e616d65c6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
