package gen_test

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type ChainTxTest struct {
	Getter     gen.InputGetter
	Key        wallet.PrivateKey
	Tx1Request gen.TxRequest
	Tx2Request gen.TxRequest
	Tx1Hash    string
	Tx2Hash    string
}

func (c ChainTxTest) Test(t *testing.T) {
	request := c.Tx1Request
	request.Getter = c.Getter
	request.KeyRing = wallet.GetSingleKeyRing(c.Key)
	tx, err := gen.Tx(request)
	if err != nil {
		t.Error(jerr.Get("error generating tx1", err))
		return
	}
	var hashString = hs.GetTxString(tx.GetHash())
	var isError bool
	if hashString != c.Tx1Hash {
		t.Error(jerr.Newf("tx1 hash %s does not match expected %s", hashString, c.Tx1Hash))
		isError = true
	}
	if testing.Verbose() || isError {
		txInfo := parse.GetTxInfo(tx)
		txInfo.Print()
	}
	if isError {
		return
	}
	request2 := c.Tx2Request
	request2.Getter = c.Getter
	request2.KeyRing = wallet.GetSingleKeyRing(c.Key)
	tx2, err := gen.Tx(request2)
	if err != nil {
		t.Error(jerr.Get("error generating tx2", err))
		return
	}
	var hashString2 = hs.GetTxString(tx2.GetHash())
	if hashString2 != c.Tx2Hash {
		t.Error(jerr.Newf("tx2 hash %s does not match expected %s", hashString2, c.Tx2Hash))
		isError = true
	}
	if testing.Verbose() || isError {
		txInfo := parse.GetTxInfo(tx2)
		fmt.Println()
		txInfo.Print()
	}
}

func TestChainTxSameAddress(t *testing.T) {
	ChainTxTest{
		Getter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosSingle25k}, test_tx.Address1pkHash),
		Key:    test_tx.GetPrivateKey(test_tx.Key1String),
		Tx1Request: gen.TxRequest{
			Outputs: []*memo.Output{
				gen.GetAddressOutput(test_tx.Address1, 5000),
			},
			Change: wallet.GetChange(test_tx.Address1),
		},
		Tx2Request: gen.TxRequest{
			Outputs: []*memo.Output{
				gen.GetAddressOutput(test_tx.Address1, 5000),
			},
			Change: wallet.GetChange(test_tx.Address1),
		},
		Tx1Hash: "e5d88a1e39b3759e0b672353007861a26a6d0ca094ad360af378512ad0353289",
		Tx2Hash: "0c8ab3aeb4da49542cb0b376dfc2cf0e03584567cdad8c74064850f6c6149c89",
	}.Test(t)
}

func TestChainTxOtherAddress(t *testing.T) {
	ChainTxTest{
		Getter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosSingle25k}, test_tx.Address1pkHash),
		Key:    test_tx.GetPrivateKey(test_tx.Key1String),
		Tx1Request: gen.TxRequest{
			Outputs: []*memo.Output{
				gen.GetAddressOutput(test_tx.Address2, 5000),
			},
			Change: wallet.GetChange(test_tx.Address1),
		},
		Tx2Request: gen.TxRequest{
			Outputs: []*memo.Output{
				gen.GetAddressOutput(test_tx.Address2, 5000),
			},
			Change: wallet.GetChange(test_tx.Address1),
		},
		Tx1Hash: "927cf1da8ca46d67382b81991206f3bae84935b39cd040e570931f73516c4554",
		Tx2Hash: "a02cd2529f4c73a37706ce457c71a67c5c775fb0163adf32440e0e73003e4eca",
	}.Test(t)
}
