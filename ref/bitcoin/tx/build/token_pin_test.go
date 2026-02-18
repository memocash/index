package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type TokenPinTest struct {
	Request  build.TokenPinRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (m TokenPinTest) Test(t *testing.T) {
	tx, err := build.TokenPin(m.Request)
	test_tx.Checker{
		Txs:      []*memo.Tx{tx},
		Error:    m.Error,
		TxHashes: m.TxHashes,
	}.Check(err, t)
}

func TestTokenPinSimple(t *testing.T) {
	TokenPinTest{
		Request: build.TokenPinRequest{
			Wallet: build.Wallet{
				Getter: gen.GetWrapper(&test_tx.TestGetter{
					UTXOs: test_tx.UtxosAddress1twoRegularWithToken,
				}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			PostTxHash:  test_tx.GenericTxHash0,
			SendTxHash:  test_tx.GenericTxHash1,
			SendTxIndex: 1,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "eb7ebb5764cc7c8399fa5e11962413c8be9a81106cec28ec72b2186a8b4142ca",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022072dcf2e096d7e828b7e10ed3b68018dd2ef4b37ea09346a74f6db724051a5d9602206d5e03ae7ee8515f5b9bcfff551bda8c9cd4bdfcf6fc2761230a2bb20e14dc4e412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000496a026d3520b158efa8e85ef8283481e000f9fb13b12599a8fa58fce12633093762ebd1cb7520ad8b36425e100db1b0bb4677dd447cf08babb493afa0fecced1e9f4d13544ad0020001be060000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
