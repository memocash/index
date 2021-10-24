package build_test

import (
	"github.com/memocash/server/ref/bitcoin/tx/build"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type PostTest struct {
	Request  build.PostRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func (tst PostTest) Test(t *testing.T) {
	txs, err := build.Post(tst.Request)
	test_tx.Checker{
		Txs:      txs,
		Error:    tst.Error,
		TxHashes: tst.TxHashes,
	}.Check(err, t)
}

func TestPostSimple(t *testing.T) {
	PostTest{
		Request: build.PostRequest{
			Wallet:  test_tx.GetAddress1WalletSingle100k(),
			Message: "Post message",
		},
		TxHashes: []test_tx.TxHash{{
			TxHash: "9f63c2ec75f4e114498fa7c46f7b5af38bf92280df669b9626c5304188a2ddad",
			TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402204cfc93f5046b9a8a54cb9b45426a2211c08999ee6556c8dc6a61980628f17ebc02207d77aa7813a4fbcb91b030426be1ab0e68e9e32a82ea8e0f778a2860379fe18f412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000116a026d020c506f7374206d657373616765c6850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
		}},
	}.Test(t)
}
