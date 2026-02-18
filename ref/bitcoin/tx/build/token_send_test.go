package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type TokenSendTest struct {
	Request  build.TokenSendRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (m TokenSendTest) Test(t *testing.T) {
	tx, err := build.TokenSend(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenSendSimple(t *testing.T) {
	TokenSendTest{
		Request: build.TokenSendRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: test_tx.UtxosAddress1twoRegularWithToken,
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			Recipient: test_tx.Address2,
			TokenHash: test_tx.SlpToken1M10,
			TokenType: memo.SlpDefaultTokenType,
			Quantity:  5,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "7aab878e665120edcd11d6dec83da7c375728b7e034a354f6147b948c886cec7",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006b483045022100e5618cc3d601059dc4e617e1f039a7d1fbe7a38b774b229aa078ea1116dce364022010c5a92a01aef9139ca006d756ce957eadae652a981e6d023f2d338e56327c7a412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402203eae0ab8ec8e99aea9ad9a1e85516a14d7e5b77bc761708a279492be1860a984022018a507da34497e14ce83e9004099c1f586301ab3aeb85a4032eac9b3e5b99d40412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff030000000000000000376a04534c500001010453454e44205ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d46608000000000000000522020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac1a060000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestTokenSendWithChange(t *testing.T) {
	TokenSendTest{
		Request: build.TokenSendRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: test_tx.UtxosAddress1twoRegularWithToken,
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			Recipient: test_tx.Address2,
			TokenHash: test_tx.SlpToken1M10,
			TokenType: memo.SlpDefaultTokenType,
			Quantity:  2,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "7809166252d916ab04ce75ae0dbf28c8738be16d750be0cdca5688c7c4334758",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006b483045022100b2c30ee055b32d3c914c3c85cb73ef68bdd2ffc08281233e9c0a205c14f7ac8d0220122eeb6d167905affdb0fc7dc9fa9016d4930ee3af65c4fec8f26c9e9cc2329a412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022007af17ed93ac898f33ac666cb57f7ff00ac21411f6b9762d00d0aa3b228966810220130a6f9fa11f2c55878ad332027accdbaded82411460aa2e64dba05e1c5372d7412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000406a04534c500001010453454e44205ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d46608000000000000000208000000000000000322020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288accd030000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestTokenSendNotEnoughTokenValue(t *testing.T) {
	TokenSendTest{
		Request: build.TokenSendRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: test_tx.UtxosAddress1twoRegularWithToken,
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			Recipient: test_tx.Address2,
			TokenHash: test_tx.SlpToken1M10,
			TokenType: memo.SlpDefaultTokenType,
			Quantity:  10,
		},
		Error: gen.NotEnoughTokenValueError,
	}.Test(t)
}

func TestTokenSendNotEnoughValue(t *testing.T) {
	TokenSendTest{
		Request: build.TokenSendRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: []memo.UTXO{test_tx.Address1InputToken},
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			Recipient: test_tx.Address2,
			TokenHash: test_tx.SlpToken1M10,
			TokenType: memo.SlpDefaultTokenType,
			Quantity:  5,
		},
		Error: gen.NotEnoughValueError,
	}.Test(t)
}
