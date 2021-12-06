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
	Error    string
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
			TxHash: "8e99e14dfb228a74bcdab15f00bb705397bd248c1f9392a5c3430a960177c28c",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100e8a29b60be5037b41093f2d34dbcfab32fee2aedf6ab706eb3e4b2b7c3cabf22022010d32b1d7eb565fa5f79f23b98563f7979a5cface82bfb063658ca21bb2666b8412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000316a04534c500001010747454e4553495304544553540a5465737420546f6b656e4c004c00010001020800000000000003e822020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac1e810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "755baf92926ff5ac976151679717ff7ae72ca496f65fdba13e15d3d6f894943d",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100dfddf981985fad09fd7269a634da4e0ef1a96891ccb69048a5256ca0c7ec53d80220172a35cd47260e8f9f9ea42dc14bebd161a0f14fb460965da758a398097e9034412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000316a04534c500001010747454e4553495304544553540a5465737420546f6b656e4c004c00010001020800000000000003e822020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac1e810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "3deeb0684df67de7e6101d9c0f21bb125b43f7273e9950267948d43ef92c1fed",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402207ff6ff1889d9712c3344e75a8f6355d4dbc57676a6df60132a0ffef3952575fe02203fecf791269aaba4e1f5041664298774b666b7f1e4b34c5b059ab6ca07da698e412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000316a04534c500001010747454e4553495304544553540a5465737420546f6b656e4c004c00010001020800000000000003e822020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac22020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac1e810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
