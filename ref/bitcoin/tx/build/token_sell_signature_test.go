package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type TokenSellSignatureTest struct {
	Request  build.TokenSellSignatureRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (m TokenSellSignatureTest) Test(t *testing.T) {
	tx, err := build.TokenSellSignature(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenSellSignatureSimple(t *testing.T) {
	TokenSellSignatureTest{
		Request: build.TokenSellSignatureRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: test_tx.UtxosAddress1twoRegularWithToken,
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			OfferTxHash: test_tx.GenericTxHash0,
			Signatures:  []script.Signature{{Sig: test_tx.SellTokenSignature, PkData: test_tx.SellTokenPkData}},
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "0b560a6e0d4bba388a3e3c1a4e7806908cb449b6ec8cc2835144a2ad57ba0e18",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402207e721e8ebc0f1bc3f880dc9e6e82582c166cc6fa289ae336f9ba683cb68e65f4022060a4f3654e3341430f38fc5cce73cc33bca4224a3a7d514e55ccabb813296b0b412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000906a026d322075cbd1eb6237093326e1fc58faa89925b113fbf900e0813428f85ee8a8ef58b1483045022100be3275298b230d8809b9c93326dfac2776f87f283c657a34bc7c65c198c9c95502206b9fb2532a1ebe1be6208d67500f506ce5fede5668b1d23a6d0de89663b8c95fc32103605c2b9b7cc8dc1063be5d7b185fb3c1fd2171bc156f4a30ef1e406789fd663177060000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
