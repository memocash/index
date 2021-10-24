package build_test

import (
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/ref/bitcoin/wallet"
	"testing"
)

type TokenSellTest struct {
	Request  build.TokenSellRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (m TokenSellTest) Test(t *testing.T) {
	tx, err := build.TokenSell(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenSellSimple(t *testing.T) {
	TokenSellTest{
		Request: build.TokenSellRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: test_tx.UtxosAddress1twoRegularWithToken,
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			InOuts: []script.InOut{script.InOutInput{
				TxHash: test_tx.GenericTxHash0,
				Index:  0,
			}, script.InOutOutput{
				Address:  test_tx.Address1,
				IsSelf:   true,
				Quantity: 1000,
			}},
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "d9e6a01a8fcf63d9e5e3c87c2aa3f662aaf3d3c7cfa0c357f60a9fc8d7eabdd8",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100e583ebc35519470922c279914fe51133c8a25f4f137e986415f59bf94f84a392022044dc12de025fc023bb6cad33d34928bacc020c64be2466ac5163108e1bde74ff412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000336a026d30512075cbd1eb6237093326e1fc58faa89925b113fbf900e0813428f85ee8a8ef58b1020000530800000000000003e8d4060000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
