package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type TokenSellAcceptTest struct {
	Request  build.TokenSellAcceptRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (m TokenSellAcceptTest) Test(t *testing.T) {
	tx, err := build.TokenSellAccept(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenSellAcceptSimple(t *testing.T) {
	TokenSellAcceptTest{
		Request: build.TokenSellAcceptRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k},
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			TokenInput: test_tx.Address2Input5Tokens1,
			PayAddress: test_tx.Address2,
			PayAmount:  1000,
			TokenHash:  test_tx.SlpToken1M10,
			TokenType:  memo.SlpDefaultTokenType,
			TokenAmt:   5,
			Signature:  test_tx.SellTokenSignature,
			PkData:     test_tx.Address2key.GetPublicKey().GetSerialized(),
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "08ee726e9ff22b4d92c332e44fa7c8a3414cff42f21505beb84accc40173b25e",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220483a12b2d56156c3f41e25bd644c5f3651e88f44aa6ae25b2b4a19ad2775de31022021c818977f18fd5a6b11b9bd51e343d064b0ae5b6818b36641bb6615370362de412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006b483045022100be3275298b230d8809b9c93326dfac2776f87f283c657a34bc7c65c198c9c95502206b9fb2532a1ebe1be6208d67500f506ce5fede5668b1d23a6d0de89663b8c95fc32102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff040000000000000000406a04534c500001010453454e44205ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d466080000000000000000080000000000000005e8030000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288acd7800100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestTokenSellAcceptWithFee(t *testing.T) {
	TokenSellAcceptTest{
		Request: build.TokenSellAcceptRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k},
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			TokenInput: test_tx.Address2Input5Tokens1,
			PayAddress: test_tx.Address2,
			PayAmount:  5000,
			TokenHash:  test_tx.SlpToken1M10,
			TokenType:  memo.SlpDefaultTokenType,
			TokenAmt:   5,
			Signature:  test_tx.SellTokenSignature,
			PkData:     test_tx.Address2key.GetPublicKey().GetSerialized(),
			FeeAddress: test_tx.Address3,
			Fee:        1000,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "3b52e0223867a80bae371312c9e2fbefcde81eba93d18bfb51e8fb45e7439e4b",
			TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100fb651dd895798f49cc7ddd0b83d399ad63469be6d374dff16e7cea3c6a8f3c3002204e1a22458b2bd2f6d9e3aca4b2635062a9d968a6aef8ead2464ce5136b2f9a75412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006b483045022100be3275298b230d8809b9c93326dfac2776f87f283c657a34bc7c65c198c9c95502206b9fb2532a1ebe1be6208d67500f506ce5fede5668b1d23a6d0de89663b8c95fc32102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff050000000000000000406a04534c500001010453454e44205ce4758425a370a68fe9a644d437b56667fad1ddf9fdf79ddfab784a6c27d46608000000000000000008000000000000000588130000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac22020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ace8030000000000001976a914bdd327223afe0556bf7edb949d17d758589ef65e88ac2d6d0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
