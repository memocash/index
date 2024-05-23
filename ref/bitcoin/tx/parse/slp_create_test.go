package parse_test

import (
	"encoding/hex"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"log"
	"testing"
)

type SlpCreateTest struct {
	PkScript   string
	Ticker     string
	Name       string
	SlpType    byte
	Decimals   uint8
	DocUrl     string
	Quantity   uint64
	BatonIndex uint32
}

func (tst SlpCreateTest) Test(t *testing.T) {
	tokenCreate := script.TokenCreate{
		Ticker:   tst.Ticker,
		Name:     tst.Name,
		SlpType:  tst.SlpType,
		Decimals: int(tst.Decimals),
		DocUrl:   tst.DocUrl,
		Quantity: tst.Quantity,
	}
	scr, err := tokenCreate.Get()
	if err != nil {
		t.Error(fmt.Errorf("error creating token create script; %w", err))
	}
	if hex.EncodeToString(scr) != tst.PkScript {
		t.Error(fmt.Errorf("error scr %x does not match expected %s", scr, tst.PkScript))
	} else if testing.Verbose() {
		log.Printf("scr %x, expected %s\n", scr, tst.PkScript)
	}
	slpCreate := parse.NewSlpCreate()
	if err := slpCreate.Parse(scr); err != nil {
		t.Error(fmt.Errorf("error parsing slp create pk script; %w", err))
	}
	if slpCreate.Ticker != tst.Ticker {
		t.Error(fmt.Errorf("slpCreate.Ticker %s does not match expected %s", slpCreate.Ticker, tst.Ticker))
	} else if testing.Verbose() {
		log.Printf("slpCreate.Ticker %s, expected %s\n", slpCreate.Ticker, tst.Ticker)
	}
	if slpCreate.Name != tst.Name {
		t.Error(fmt.Errorf("slpCreate.Name %s does not match expected %s", slpCreate.Name, tst.Name))
	} else if testing.Verbose() {
		log.Printf("slpCreate.Name %s, expected %s\n", slpCreate.Name, tst.Name)
	}
	if slpCreate.TokenType != tst.SlpType {
		t.Error(fmt.Errorf("slpCreate.SlpType %s does not match expected %s",
			memo.SlpTypeByteString(slpCreate.TokenType), memo.SlpTypeByteString(tst.SlpType)))
	} else if testing.Verbose() {
		log.Printf("slpCreate.SlpType %s, expected %s\n",
			memo.SlpTypeByteString(slpCreate.TokenType), memo.SlpTypeByteString(tst.SlpType))
	}
	if slpCreate.Decimals != tst.Decimals {
		t.Error(fmt.Errorf("slpCreate.Decimals %d does not match expected %d", slpCreate.Decimals, tst.Decimals))
	} else if testing.Verbose() {
		log.Printf("slpCreate.Decimals %d, expected %d\n", slpCreate.Decimals, tst.Decimals)
	}
	if slpCreate.DocUrl != tst.DocUrl {
		t.Error(fmt.Errorf("slpCreate.DocUrl %s does not match expected %s", slpCreate.DocUrl, tst.DocUrl))
	} else if testing.Verbose() {
		log.Printf("slpCreate.DocUrl %s, expected %s\n", slpCreate.DocUrl, tst.DocUrl)
	}
	if slpCreate.Quantity != tst.Quantity {
		t.Error(fmt.Errorf("slpCreate.Quantity %d does not match expected %d", slpCreate.Quantity, tst.Quantity))
	} else if testing.Verbose() {
		log.Printf("slpCreate.Quantity %d, expected %d\n", slpCreate.Quantity, tst.Quantity)
	}
	if slpCreate.BatonIndex != tst.BatonIndex {
		t.Error(fmt.Errorf("slpCreate.BatonIndex %d does not match expected %d", slpCreate.BatonIndex, tst.BatonIndex))
	} else if testing.Verbose() {
		log.Printf("slpCreate.BatonIndex %d, expected %d\n", slpCreate.BatonIndex, tst.BatonIndex)
	}
}

const (
	SlpCreateDefaultScript  = "6a04534c500001010747454e455349530254540a5465737420546f6b656e1268747470733a2f2f746f6b656e2e746573744c0001020102080000000000002710"
	SlpCreateNftGroupScript = "6a04534c500001810747454e455349530254540a5465737420546f6b656e1268747470733a2f2f746f6b656e2e746573744c0001050102080000000000000032"
	SlpCreateNftChildScript = "6a04534c500001410747454e455349530254540a5465737420546f6b656e1268747470733a2f2f746f6b656e2e746573744c0001004c000800000000000005dc"
)

func TestSlpCreateDefault(t *testing.T) {
	SlpCreateTest{
		PkScript:   SlpCreateDefaultScript,
		Ticker:     test_tx.TestTokenTicker,
		Name:       test_tx.TestTokenName,
		SlpType:    memo.SlpDefaultTokenType,
		Decimals:   2,
		DocUrl:     test_tx.TestTokenDocUrl,
		Quantity:   10000,
		BatonIndex: 2,
	}.Test(t)
}

func TestSlpCreateNftGroup(t *testing.T) {
	SlpCreateTest{
		PkScript:   SlpCreateNftGroupScript,
		Ticker:     test_tx.TestTokenTicker,
		Name:       test_tx.TestTokenName,
		SlpType:    memo.SlpNftGroupTokenType,
		Decimals:   5,
		DocUrl:     test_tx.TestTokenDocUrl,
		Quantity:   50,
		BatonIndex: 2,
	}.Test(t)
}

func TestSlpCreateNftChild(t *testing.T) {
	SlpCreateTest{
		PkScript: SlpCreateNftChildScript,
		Ticker:   test_tx.TestTokenTicker,
		Name:     test_tx.TestTokenName,
		SlpType:  memo.SlpNftChildTokenType,
		Decimals: 0,
		DocUrl:   test_tx.TestTokenDocUrl,
		Quantity: 1500,
	}.Test(t)
}
