package parse_test

import (
	"bytes"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"testing"
)

type SlpMintTest struct {
	PkScript   string
	SlpType    uint16
	TokenHash  []byte
	Quantity   uint64
	BatonIndex uint32
}

func (tst SlpMintTest) Test(t *testing.T) {
	tokenMint := script.TokenMint{
		TokenHash: tst.TokenHash,
		TokenType: byte(tst.SlpType),
		Quantity:  tst.Quantity,
	}
	scr, err := tokenMint.Get()
	if err != nil {
		t.Error(jerr.Get("error creating token mint script", err))
	}
	if hex.EncodeToString(scr) != tst.PkScript {
		t.Error(jerr.Newf("error scr %x does not match expected %s", scr, tst.PkScript))
	} else if testing.Verbose() {
		jlog.Logf("scr %x, expected %s\n", scr, tst.PkScript)
	}
	slpMint := parse.NewSlpMint()
	if err := slpMint.Parse(scr); err != nil {
		t.Error(jerr.Get("error parsing slp create pk script", err))
	}
	if slpMint.TokenType != tst.SlpType {
		t.Error(jerr.Newf("slpMint.SlpType %s does not match expected %s",
			memo.SlpTypeString(slpMint.TokenType), memo.SlpTypeString(tst.SlpType)))
	} else if testing.Verbose() {
		jlog.Logf("slpMint.SlpType %s, expected %s\n",
			memo.SlpTypeString(slpMint.TokenType), memo.SlpTypeString(tst.SlpType))
	}
	if !bytes.Equal(slpMint.TokenHash, tst.TokenHash) {
		t.Error(jerr.Newf("slpMint.TokenHash %x does not match expected %x", slpMint.TokenHash, tst.TokenHash))
	} else if testing.Verbose() {
		jlog.Logf("slpMint.TokenHash %x, expected %x\n", slpMint.TokenHash, tst.TokenHash)
	}
	if slpMint.Quantity != tst.Quantity {
		t.Error(jerr.Newf("slpMint.Quantity %d does not match expected %d", slpMint.Quantity, tst.Quantity))
	} else if testing.Verbose() {
		jlog.Logf("slpMint.Quantity %d, expected %d\n", slpMint.Quantity, tst.Quantity)
	}
	if slpMint.BatonIndex != tst.BatonIndex {
		t.Error(jerr.Newf("slpMint.BatonIndex %d does not match expected %d", slpMint.BatonIndex, tst.BatonIndex))
	} else if testing.Verbose() {
		jlog.Logf("slpMint.BatonIndex %d, expected %d\n", slpMint.BatonIndex, tst.BatonIndex)
	}
}

const (
	SlpMintDefaultScript  = "6a04534c50000101044d494e5420b158efa8e85ef8283481e000f9fb13b12599a8fa58fce12633093762ebd1cb750102080000000000002710"
	SlpMintNftGroupScript = "6a04534c50000181044d494e5420ad8b36425e100db1b0bb4677dd447cf08babb493afa0fecced1e9f4d13544ad00102080000000000000032"
	SlpMintNftChildScript = "6a04534c50000141044d494e5420e0a9936a36780efa0e50e30cb466e8077c70623cba95a28e3b2125754c98aab701020800000000000005dc"
)

func TestSlpMintDefault(t *testing.T) {
	SlpMintTest{
		PkScript:   SlpMintDefaultScript,
		SlpType:    memo.SlpDefaultTokenType,
		TokenHash:  test_tx.GenericTxHash0,
		Quantity:   10000,
		BatonIndex: 2,
	}.Test(t)
}

func TestSlpMintNftGroup(t *testing.T) {
	SlpMintTest{
		PkScript:   SlpMintNftGroupScript,
		SlpType:    memo.SlpNftGroupTokenType,
		TokenHash:  test_tx.GenericTxHash1,
		Quantity:   50,
		BatonIndex: 2,
	}.Test(t)
}

func TestSlpMintNftChild(t *testing.T) {
	SlpMintTest{
		PkScript:   SlpMintNftChildScript,
		SlpType:    memo.SlpNftChildTokenType,
		TokenHash:  test_tx.GenericTxHash2,
		Quantity:   1500,
		BatonIndex: 2,
	}.Test(t)
}
