package gen_test

import (
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"testing"
)

type OutputTest struct {
	Address wallet.Address
	Amount  int64
	Script  string
	Type    memo.OutputType
}

var outputTest0p2pkh = OutputTest{
	Address: test_tx.Address1,
	Amount:  1000,
	Script:  "76a914fc393e225549da044ed2c0011fd6c8a799806b6288ac",
	Type:    memo.OutputTypeP2PKH,
}

var outputTest1p2sh = OutputTest{
	Address: test_tx.AddressP2sh1,
	Amount:  1000,
	Script:  "a914dd763c90ae1a5677d925c680673bba0a5e28740587",
	Type:    memo.OutputTypeP2SH,
}

var tests = []OutputTest{
	outputTest0p2pkh,
	outputTest1p2sh,
}

func TestGetOutput(t *testing.T) {
	for _, tst := range tests {
		output := gen.GetAddressOutput(tst.Address, tst.Amount)
		jlog.Logf("%s - output.Script.Type(): %s\n", tst.Address.GetEncoded(), output.Script.Type())
		pkScript, err := output.GetPkScript()
		if err != nil {
			t.Error(jerr.Get("error getting pk script", err))
			continue
		}
		if tst.Script != hex.EncodeToString(pkScript) {
			t.Error(jerr.Newf("pkScript (%x) does not match expected (%x)", pkScript, tst.Script))
		}
		if output.GetType() != tst.Type {
			t.Error(jerr.Newf("output type (%s) does not match expected (%s)", output.GetType(), tst.Type))
		}
	}
}
