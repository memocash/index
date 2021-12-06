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
	Error    string
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
			TxHash: "24b2e37ab5b8ec90988c0e3c7e79282655c79b999c2260f5d86c6b10ed85e471",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006a47304402205cfbd34cf18645234cdb446a6cc8d37c552c90b83ab2e116dc66bc333f11463702206e655e40746bb44111c06554579050cfccb35b47fc0d2152292ce0f83cf662a6412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b48304502210097c9fa5f3abbdfb9ae403c1353b4fb7a7e442156c5645ac5b4abd716cfce65590220073846d42d64b0c4abf34afd933040e79204297ca9e4745612183637d1d76ae0412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff030000000000000000376a04534c500001010453454e44205ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d46608000000000000000522020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac1a060000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "68736e6bf257b4a21f27e6d25d8816cd72259b03cc21f9b9c6f42e088a39fb49",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006b48304502210081a8de4d2474ebbda6530e6773fd8d3a92125f52a2c868865213a8bee163615402200c266c0b2d2b275516e05c6a157d02c6267f856f440d00df11b81eab0410c93e412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022001916e4d35a347fbe0cd31985118c9f0d6c9378ed8c53d8d25c28216f6f9519602201dfa35fc81518bb41ff8ad0349ab25707bdd260843a3078a9d379380127163e9412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff040000000000000000406a04534c500001010453454e44205ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d46608000000000000000208000000000000000322020000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288accd030000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
		Error: gen.NotEnoughTokenValueErrorText,
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
		Error: gen.NotEnoughValueErrorText,
	}.Test(t)
}
