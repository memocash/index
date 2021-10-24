package wallet_test

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/wallet"
	"testing"
)

var testSlps = []struct {
	InputAddress  string
	SlpAddress    string
	LegacyAddress string
}{{
	InputAddress:  "1QCBiyfwdjXDsHghBEr5U2KxUpM2BmmJVt",
	SlpAddress:    "qrlxs6um926cng7txd5dqgs3egdfhz92gg0e56pvw6",
	LegacyAddress: "1QCBiyfwdjXDsHghBEr5U2KxUpM2BmmJVt",
}, {
	InputAddress:  "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
	SlpAddress:    "prpmwwdwgz5d3e70xtmckzvccq35pznxgsper7gdmw",
	LegacyAddress: "3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
}, {
	InputAddress:  "bitcoincash:3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
	SlpAddress:    "",
	LegacyAddress: "",
}, {
	InputAddress:  "bitcoin:3KXsBfjyr7StP92q1bse6oaTeG3JUjxteL",
	SlpAddress:    "",
	LegacyAddress: "",
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
		addr := wallet.GetAddressFromString(testSlp.InputAddress)
		slpAddress := addr.GetSlpAddrString()
		if slpAddress != testSlp.SlpAddress {
			t.Error(jerr.Newf("slp address (%s) doesn't match expected (%s)", slpAddress, testSlp.SlpAddress))
		}
		legacyAddress := addr.GetEncoded()
		if legacyAddress != testSlp.LegacyAddress {
			t.Error(jerr.Newf("legacy address (%s) doesn't match expected (%s)", legacyAddress, testSlp.LegacyAddress))
		}
	}
}
