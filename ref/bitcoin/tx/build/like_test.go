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
	Error    string
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
			TxHash: "d4c01b19b50f249d04779bd7acc510026fb215bac9eaa61c4fe56f6a3693f8ca",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220203b4b23d15054ad92ecccb99f4d2d82b7da3982950b0b3a638dda6d3004847c02201bd7ba83905277b4995f0d6564dad35bd2edd583fe651d2be65d58efd55ba325412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "d8419953de272e264e662ebfde6f199a12b1b7772f65854e67c88d2636a79b2c",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100fdfaca1281312d5018b6a933b4b45c5cdc90cd2031d547c67487c93752afaf10022014b13fc6e77c77411eb4ead85031173343f87f000af7eb7854c3eaaaa12d388c412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff030000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2e8030000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588aca8810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
			TxHash: "aff0c0c028a6e539829b8000f9f610ea6e81dc844c6539def157f0d677ac74a4",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100e930c832bb403a59dfc88fcca75c83852f29b67e42a9f7856a4ee06f09cdaae302200815fbebbd853d2843413c297bcd840a9c204a61963ec584748d41b02d4292d8412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff030000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2e80300000000000017a914dd763c90ae1a5677d925c680673bba0a5e28740587aa810100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
