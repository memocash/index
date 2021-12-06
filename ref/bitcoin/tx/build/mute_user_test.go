package build_test

import (
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type MuteTest struct {
	Request  build.MuteUserRequest
	Error    string
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
			TxHash: "2b297840d2aa91df997388a10409961156663001dd3599996abdc6ff80767a65",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100be7216bc619a8517fee8293b85e4d08f2dea50af670ffe6f6dead832fecedbdd02205771c4f5bcdfbf1073b87951e2a75e8fef7886e4585c96942ee4cbe88bc71204412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d16140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "1016ecca958fcfdfa7070b01427c375aefd7ee4c9f89180efd7ecf48345cb6e4",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220234a5b82dd12627294eef6c99b98070483b2858e036b4d50dcd4b34e04fe4274022038ba7162094c587e4bd388f9c3843e7a62e36ae96f6c3e8fe6dd17d2a9e7681f412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000196a026d17140d4cd6490ddf863bbdf5c34d8ef1aebfd45c2105be850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
