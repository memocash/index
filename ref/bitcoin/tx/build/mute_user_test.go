package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type MuteTest struct {
	Request  build.MuteUserRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst MuteTest) Test(t *testing.T) {
	txs, err := build.MuteUser(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestMuteSimple(t *testing.T) {
	MuteTest{
		Request: build.MuteUserRequest{
			Wallet:     test_tx.GetAddress1WalletSingle100k(),
			MutePkHash: test_tx.Address2pkHash,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "3e38f47b4b3722801b10e76bed257a655cae816a46d1260694c9eaba94861798",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402206b6fe672cb34ab09aac26e0d0cc3479d5994546eb3beee653ad11d0d365538d4022072b6ba68eeb5617fee2d0f79f23c2cf552e3d4c01c32ed1b107b56751cf95d5c412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d16140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestUnmuteSimple(t *testing.T) {
	MuteTest{
		Request: build.MuteUserRequest{
			Wallet:     test_tx.GetAddress1WalletSingle100k(),
			MutePkHash: test_tx.Address2pkHash,
			Unmute:     true,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "0de88978287fa7d117c7f336ce4679729a0724f491ffa3800f6159f6ec75ff45",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022005a9d2062609b52f8d243cfc47d5c05053238a4fd6fe1c5541712c1cc30c9111022068e063db29ace31b8107947b772db34966181fccdd1e178862ddeb7b79167e70412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d17140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
