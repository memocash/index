package wallet_test

import (
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/wallet"
	"testing"
)

var testSlps = []struct {
	InputAddress  string
	SlpAddress    string
	LegacyAddress string
	Error         bool
}{{
	InputAddress:  "1QCBiyfwdjXDsHghBEr5U2KxUpM2BmmJVt",
	SlpAddress:    "qrlxs6um926cng7txd5dqgs3egdfhz92gg0e56pvw6",
	LegacyAddress: "1QCBiyfwdjXDsHghBEr5U2KxUpM2BmmJVt",
}, {
	InputAddress:  "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
	SlpAddress:    "prpmwwdwgz5d3e70xtmckzvccq35pznxgsper7gdmw",
	LegacyAddress: "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
}, {
	InputAddress: "bitcoincash:3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
	Error:        true,
}, {
	InputAddress: "bitcoin:3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
	Error:        true,
}, {
	InputAddress:  "bitcoincash:ppfuzf5f52ypwm0jrnnfqtwafrtfx0zgcshtux8zk3",
	SlpAddress:    "ppfuzf5f52ypwm0jrnnfqtwafrtfx0zgcsmshajzg0",
	LegacyAddress: "39KsPoNVQHmGvoqEudignH2QbFoD4E3YqK",
}, {
	InputAddress:  "prpmwwdwgz5d3e70xtmckzvccq35pznxgsdzg9ad9s",
	SlpAddress:    "prpmwwdwgz5d3e70xtmckzvccq35pznxgsper7gdmw",
	LegacyAddress: "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
}, {
	InputAddress:  "bitcoincash:prpmwwdwgz5d3e70xtmckzvccq35pznxgsdzg9ad9s",
	SlpAddress:    "prpmwwdwgz5d3e70xtmckzvccq35pznxgsper7gdmw",
	LegacyAddress: "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
}, {
	InputAddress:  "prpmwwdwgz5d3e70xtmckzvccq35pznxgsper7gdmw",
	SlpAddress:    "prpmwwdwgz5d3e70xtmckzvccq35pznxgsper7gdmw",
	LegacyAddress: "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
}, {
	InputAddress:  "simpleledger:ppfuzf5f52ypwm0jrnnfqtwafrtfx0zgcsmshajzg0",
	SlpAddress:    "ppfuzf5f52ypwm0jrnnfqtwafrtfx0zgcsmshajzg0",
	LegacyAddress: "39KsPoNVQHmGvoqEudignH2QbFoD4E3YqK",
}}

func TestSlpAddress(t *testing.T) {
	for _, testSlp := range testSlps {
		addr, err := wallet.GetAddressFromStringErr(testSlp.InputAddress)
		if testSlp.Error {
			if err == nil {
				t.Error(jerr.Newf("slp address expected error, but none (%s)", testSlp.SlpAddress))
			}
			continue
		}
		if err != nil {
			t.Error(jerr.Getf(err, "slp address unexpected error (%s)", testSlp.SlpAddress))
			continue
		}
		slpAddress := addr.GetSlpAddrString()
		if slpAddress != testSlp.SlpAddress {
			t.Error(jerr.Newf("slp address (%s) doesn't match expected (%s)", slpAddress, testSlp.SlpAddress))
			continue
		}
		legacyAddress := addr.GetEncoded()
		if legacyAddress != testSlp.LegacyAddress {
			t.Error(jerr.Newf("legacy address (%s) doesn't match expected (%s)", legacyAddress, testSlp.LegacyAddress))
			continue
		}
	}
}

const RawTxString = "01000000013876e935aa8dc4fbacb606dec3edd7d7f6dc0501eed63e01eb63398a88d866e5000000006b483045022100db32b6fabc99a481c4cb482d1dd5303048522ce619f98f380bc9d5968e4c9ed802201aeec65a10926daf904c76da6c7d3e9ec210b27604d640219cfad11be29c6fdd4121023427fa5ba55fb541c76ff5d20b9d637bd7fcb08327735ce230d614df43c9c36fffffffff06383f1b00000000001976a9144c42f071ec26b7253b2f7416849768ec41b6fd8888ace803000000000000695321037b227478223a207b2276657273696f6e223a2022312e31222c20226e65775f72210365636569707473223a205b7b2273656e645f72656365697074223a207b22636f2103696e223a20302c2022616d6f756e74223a2032373938312c202274617267657453aee803000000000000695321035f696e766f696365223a207b22696e646578223a203130323137313732353633210332382c20227075625f6b65795f68617368223a202264386261373932353262612103333530393864386566643664373732383063646562303433626130303564303853aee80300000000000069532103623765313437356233656136656138653335346439227d7d7d5d2c2022707265210376696f75735f7265636569707473223a205b7b227265636569766572223a207b2103227369676e6572223a202230333334613062383733353461636236323538393353aee803000000000000695321036332343431363461613637623336643232373231653663383139313631393633210334393365643939353534386364227d2c2022726563656970745f72656665726521036e6365223a20226266303166636635633161346130653266323762353532613753aee8030000000000004752210363306434613337623530643937376364663365346665383566643037356164322103326439393132375f31227d5d7d7d00000000000000000000000000000000000052ae00000000"

func TestPkScriptAddress(t *testing.T) {
	t.SkipNow()
	rawTx, err := hex.DecodeString(RawTxString)
	if err != nil {
		t.Error(jerr.Get("error parsing raw tx", err))
		return
	}
	msgTx, err := memo.GetMsgFromRaw(rawTx)
	if err != nil {
		t.Error(jerr.Get("error getting message from raw", err))
		return
	}
	for i, txOut := range msgTx.TxOut {
		address, err := wallet.GetAddressFromPkScript(txOut.PkScript)
		if err != nil {
			t.Error(jerr.Get("error getting address from pk script", err))
			return
		}
		jlog.Logf("i: %d, address: %s\n", i, address.GetEncoded())
	}
}
