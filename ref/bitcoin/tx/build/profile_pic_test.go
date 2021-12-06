package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type ProfilePicTest struct {
	Request  build.ProfilePicRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (tst ProfilePicTest) Test(t *testing.T) {
	txs, err := build.ProfilePic(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestProfilePicSimple(t *testing.T) {
	ProfilePicTest{
		Request: build.ProfilePicRequest{
			Wallet: test_tx.GetAddress1WalletSingle100k(),
			Url:    "https://i.imgur.com/2mZbC7Z.png",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "a7e2952d5e213108dcc186330d55e6715a0ea72916c9d8ab9dc3da1bb8192a31",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100ca8bc7de166dd4b1fef40e3f4da70cf09c8a80ee012c73dcca0eb84188e0b7e30220126a0e0b34ad0f2d7c78253966d4f12cc5cefa408f2a741e86e593907e8087cf412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000246a026d0a1f68747470733a2f2f692e696d6775722e636f6d2f326d5a6243375a2e706e67b3850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
