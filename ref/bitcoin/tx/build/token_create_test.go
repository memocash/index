package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type TokenCreateTest struct {
	Request  build.TokenCreateRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (m TokenCreateTest) Test(t *testing.T) {
	tx, err := build.TokenCreate(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenCreateSimple(t *testing.T) {
	TokenCreateTest{
		Request: build.TokenCreateRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			SlpType:  memo.SlpDefaultTokenType,
			Ticker:   "TEST",
			Name:     "Test Token",
			Quantity: 1000,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "d559f6b7de53fb52ec3ea2a51f0c91e0131604c8f72beb7bad133181ad0746e4",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100cc31730103a709d4967ba2175f263f5fe193d02d1015e6bd463e324e3813d00e02206f8eb69603099879b1b52914aad9d3695aa446096415a6eff180020191933cb6412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000316a04534c500001010747454e4553495304544553540a5465737420546f6b656e4c004c00010001020800000000000003e822020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac1e810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestTokenCreateDifferentTokenAddress(t *testing.T) {
	TokenCreateTest{
		Request: build.TokenCreateRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			SlpType:      memo.SlpDefaultTokenType,
			Ticker:       "TEST",
			Name:         "Test Token",
			Quantity:     1000,
			TokenAddress: test_tx.Address2,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "240eb844ddfb98e63aad85517cfb4cbb58780a04fe13e8360741116d7ffb84ee",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220321f1138a85159963afde340286fc090e53ba31c9255bb125b8476edfe9c6bde0220159121ad19cf68f4132df20558e9fb6fbe3f2ae1ac33fa270de6e775e9e07cad412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000316a04534c500001010747454e4553495304544553540a5465737420546f6b656e4c004c00010001020800000000000003e822020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac1e810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestTokenCreateDifferentBatonAddress(t *testing.T) {
	TokenCreateTest{
		Request: build.TokenCreateRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			SlpType:      memo.SlpDefaultTokenType,
			Ticker:       "TEST",
			Name:         "Test Token",
			Quantity:     1000,
			BatonAddress: test_tx.Address2,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "c1e64b8d7ae6191026ed4d25846b1f237bb96ea4220d11f89382455f76b993f2",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022048899aaea721f11c48b0e6b5831136531200002489c967eaf7b68a47cecea51902201fe7e459196e9a47bf5c0e7672a9da31323192fac5c9351feec7bb2150c2f6f3412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000316a04534c500001010747454e4553495304544553540a5465737420546f6b656e4c004c00010001020800000000000003e822020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac22020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac1e810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
