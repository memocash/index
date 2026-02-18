package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type ProfilePicTest struct {
	Request  build.ProfilePicRequest
	Error    error
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
			TxHash: "89b925c28dbf4184401191ed034fbf7664db9444376e89a7ad55c452d3cce384",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100f7d0677ec7198bd7a3ec90dfe6d8916a74efa4051c7c0a9ca802f3b4d652b4cd02200b45301747f6fed2588994038f5dfb869f8405d9acddc29987debcd70d76972b412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000246a026d0a1f68747470733a2f2f692e696d6775722e636f6d2f326d5a6243375a2e706e67b3850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
