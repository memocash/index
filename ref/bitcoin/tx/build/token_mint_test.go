package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type TokenMintTest struct {
	Request  build.TokenMintRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (m TokenMintTest) Test(t *testing.T) {
	tx, err := build.TokenMint(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenMintSimple(t *testing.T) {
	TokenMintTest{
		Request: build.TokenMintRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			BatonAddress: test_tx.Address1,
			TokenAddress: test_tx.Address1,
			Baton:        test_tx.Address1InputTokenBaton,
			TokenHash:    test_tx.HashEmptyTx,
			TokenType:    memo.SlpDefaultTokenType,
			Quantity:     1000,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "8ccb2e912ec1d40fbfa08dde7814956a87f44ee539e441346343ad2fe349ac0b",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006a4730440220124a3530c3019fee92db04c29c4d52c780e7c7717f6480e9799cc02c5511ee4c02204e41826446bbf5cdef5c840d27332d24bc924d29701a17fdf08165dcd82ab0d6412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402206b3015ec076779d97087797085e6240e28da37936620b505dfe7da2decf0fcb102204ce9671f4c71cb66dc929082962f3091f6c956d0cf8ed258095e3fc910957ad7412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000396a04534c50000101044d494e5420d21633ba23f70118185227be58a63527675641ad37967e2aa461559f577aec4301020800000000000003e822020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288aca4820100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
