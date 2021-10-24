package build_test

import (
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/ref/bitcoin/wallet"
	"testing"
)

type TokenMintTest struct {
	Request  build.TokenMintRequest
	Error    string
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
			TxHash: "8e5f1c40d027520b068662c3d012be72e2a3f6b95344abe0f830c20b5f0b8a39",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006a47304402202f572be6237980df97c5aac00691f0917515e9f4c5f2bde9c7c0f3bbea3f472902204d9c0a89047007234d358907e54ca33ac5bb09b3ecfb96be8b0c9db1beb08e5e412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402207993bef77ff67bad03ed7f7339bc56268a421cf938ca49e34cdfe3efeae07252022007861275f7f42606de846f83cbd258b4f59f57189a0838096610a2138c606371412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000396a04534c50000101044d494e5420d21633ba23f70118185227be58a63527675641ad37967e2aa461559f577aec4301020800000000000003e822020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288aca4820100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
