package build_test

import (
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type AliasTest struct {
	Request  build.AliasRequest
	Error    string
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
			TxHash: "86a5324fe5a116ff22b624e743a424e316a240ac699b0cdb83c1727911551824",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402204a219a6976f00548fe0a3a2915d3e3dc536bc6f8e5ce291e37f31f2ddd616373022073e7b5025f29a4ffa7975289c0dc00a655701f213e59a08b30516648593fc2f8412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0200000000000000001f6a026d2614fc393e225549da044ed2c0011fd6c8a799806b6205416c696173b8850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
