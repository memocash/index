package gen_test

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type Test struct {
	Req  gen.TxRequest
	Sign bool
	Hash string
	Err  string
}

func (tst Test) Test(t *testing.T) {
	var memoTx *memo.Tx
	var err error
	if tst.Sign {
		memoTx, err = gen.Tx(tst.Req)
	} else {
		memoTx, err = gen.TxUnsigned(tst.Req)
	}
	if err != nil {
		if tst.Err != "" && jerr.HasError(err, tst.Err) {
			if testing.Verbose() {
				jlog.Logf("Tx has expected error - '%s'\n", tst.Err)
				jerr.Create(err).Print()
			}
		} else {
			t.Error(jerr.Getf(err, "error generating tx (%s)", tst.Err))
		}
	} else {
		var isError bool
		if ! bytes.Equal(memoTx.GetHash(), test_tx.GetHashBytes(tst.Hash)) {
			t.Error(jerr.Newf("tx hash (%s) does not match expected (%s)",
				memoTx.MsgTx.TxHash().String(), tst.Hash))
			isError = true
		}
		if testing.Verbose() || isError {
			txInfo := parse.GetTxInfo(memoTx)
			txInfo.Print()
		}
	}
}

func TestEmpty(t *testing.T) {
	Test{
		Req: gen.TxRequest{},
		Err: gen.NotEnoughValueErrorText,
	}.Test(t)
}

func TestUnsignedLikeWithTip(t *testing.T) {
	Test{
		Hash: "0ab8c792857c70a44f6dc54322910dbf836b7fd6919a3a77c7c4843d98a61222",
		Req: gen.TxRequest{
			Getter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosSingle25k}, test_tx.Address1pkHash),
			Outputs: []*memo.Output{
				gen.GetAddressOutput(test_tx.Address1, 5000),
			},
			Change: wallet.GetChange(test_tx.Address1),
		},
	}.Test(t)
}

func TestUnsignedLike(t *testing.T) {
	Test{
		Hash: "0c6a534b12f266da43d3b9d151107ec748c75326fd91c5ff494a137db6382609",
		Req: gen.TxRequest{
			Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosSingle25k}, test_tx.Address1pkHash),
			Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
			Change:  wallet.GetChange(test_tx.Address1),
		},
	}.Test(t)
}

func TestSignedLike(t *testing.T) {
	Test{
		Hash: "d4c01b19b50f249d04779bd7acc510026fb215bac9eaa61c4fe56f6a3693f8ca",
		Sign: true,
		Req: gen.TxRequest{
			KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key1String)),
			Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
			Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
			Change:  wallet.GetChange(test_tx.Address1),
		},
	}.Test(t)
}

func TestMultiInput(t *testing.T) {
	Test{
		Hash: "1560058808f3ad3a3fe1e54939ce66459c4e98f7311464b3cbffae3173046da1",
		Sign: true,
		Req: gen.TxRequest{
			KeyRing: wallet.KeyRing{Keys: []wallet.PrivateKey{
				test_tx.GetPrivateKey(test_tx.Key1String),
				test_tx.GetPrivateKey(test_tx.Key2String),
			}},
			Getter: gen.GetWrapperMultiKey(&test_tx.TestGetter{UTXOs: []memo.UTXO{
				test_tx.Address1InputUtxo100k,
				test_tx.Address2InputUtxo100k,
			}}, [][]byte{
				test_tx.Address1pkHash,
				test_tx.Address2pkHash,
			}),
			Outputs: []*memo.Output{gen.GetAddressOutput(test_tx.Address1, 150000)},
			Change:  wallet.GetChange(test_tx.Address2),
		},
	}.Test(t)
}
