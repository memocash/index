package build_test

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type ReplyTest struct {
	Request  build.ReplyRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func (tst ReplyTest) Test(t *testing.T) {
	txs, err := build.Reply(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestReplySimple(t *testing.T) {
	ReplyTest{
		Request: build.ReplyRequest{
			Wallet:  test_tx.GetAddress1WalletSingle100k(),
			TxHash:  test_tx.HashEmptyTx,
			Message: "Reply message",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "c357a49ce98be4faf670a010f94c84f29862a92b4139712b4ccc1e185d6c072d",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100faacba022b27e16ed82c5979731139836e364f7575ff4c3686f6d9e1e0c8efc802203ba7b18f18cb3d5f8f1cacf6e9be98c744605acd842b4d29722b540a2f8ddf04412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000336a026d032043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d20d5265706c79206d657373616765a4850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}

func TestReplyDustLimit(t *testing.T) {
	ReplyTest{
		Request: build.ReplyRequest{
			Wallet: build.Wallet{
				Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo861}}, test_tx.Address1pkHash),
				Address: test_tx.Address1,
				KeyRing: wallet.GetSingleKeyRing(test_tx.Address1key),
			},
			TxHash:  test_tx.HashEmptyTx,
			Message: "ive been sending 5k 6k or 8k honks in here but it will only add me 200sats but honk tokens are missing",
		},
		Error: gen.NotEnoughValueError,
	}.Test(t)
}
