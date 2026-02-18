package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type LikeTest struct {
	Request  build.LikeRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst LikeTest) Test(t *testing.T) {
	txs, err := build.Like(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestLikeSimple(t *testing.T) {
	LikeTest{
		Request: build.LikeRequest{
			Wallet: test_tx.GetAddress1WalletSingle100k(),
			TxHash: test_tx.HashEmptyTx,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "985f3841672cfc1a61a547f9f98e0d618f4b5d90f4b6ed3d68597075c09ba9f4",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100ec3512456cdf33d5fb57ab7c46896966235d38507b4c89b009a58882013f126902203a3c5b963f28b2166bfcea3ed77a0eda58ba80bb4e1cc13d889c268e0addb0a9412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestLikeSimpleTip(t *testing.T) {
	LikeTest{
		Request: build.LikeRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			TxHash:     test_tx.HashEmptyTx,
			TipAddress: test_tx.Address2,
			Tip:        1000,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "3433fa412e5356c34c3095de102f42085bc587d402d7c83e89624f9f6a3fc450",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100b5ced39bd40a940e8d502356d1c699c51f95d02dd3d7a7a4ab8f53c0bbd65a5b02205949cab739c9f3a1bbba43487cfab1f5d8acbaef6f9831d4d46e84dbd8ad1097412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff030000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2e8030000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588aca8810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestLikeP2shTip(t *testing.T) {
	LikeTest{
		Request: build.LikeRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			TxHash:     test_tx.HashEmptyTx,
			TipAddress: test_tx.AddressP2sh1,
			Tip:        1000,
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "b25a2335322fe0a91b1e02578604f0cd256d413ef2365555c87ae08f2e9d8242",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100c83ce8874279a104a5a4d1c8af8522ac8c7dbd47cf6a5b7c8fa325ada6631b3a0220746f315bfad7323306f607fceb8abd5cd73c5bb4293b5927ea136b939c772270412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff030000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2e80300000000000017a914dd763c90ae1a5677d925c680673bba0a5e28740587aa810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
