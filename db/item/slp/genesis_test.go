package slp_test

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

func TestGenesis(t *testing.T) {
	txHash, _ := chainhash.NewHash(test_tx.GenericTxHash0)
	var genesis = &slp.Genesis{
		TxHash:     *txHash,
		Addr:       test_tx.Address1.GetAddr(),
		TokenType:  memo.SlpDefaultTokenType,
		Decimals:   8,
		BatonIndex: 2,
		Quantity:   100000,
		DocHash:    [32]byte{},
		Ticker:     "TEST",
		Name:       "Test Token Name",
		DocUrl:     "https://example.com",
	}
	data := genesis.Serialize()
	var genesis2 slp.Genesis
	genesis2.SetUid(genesis.GetUid())
	genesis2.Deserialize(data)
	if genesis2.TxHash != genesis.TxHash {
		t.Error("TxHash not equal")
	} else if genesis2.Addr != genesis.Addr {
		t.Error("Addr not equal")
	} else if genesis2.TokenType != genesis.TokenType {
		t.Error("TokenType not equal")
	} else if genesis2.Decimals != genesis.Decimals {
		t.Error("Decimals not equal")
	} else if genesis2.BatonIndex != genesis.BatonIndex {
		t.Error("BatonIndex not equal")
	} else if genesis2.Quantity != genesis.Quantity {
		t.Error("Quantity not equal")
	} else if genesis2.DocHash != genesis.DocHash {
		t.Error("DocHash not equal")
	} else if genesis2.Ticker != genesis.Ticker {
		t.Error("Ticker not equal")
	} else if genesis2.Name != genesis.Name {
		t.Error("Name not equal")
	} else if genesis2.DocUrl != genesis.DocUrl {
		t.Error("DocUrl not equal")
	}
}
