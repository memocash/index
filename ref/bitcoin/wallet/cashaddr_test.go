package wallet

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/jchavannes/bchutil"
)

// Vectors from the CashAddr spec (spec/cashaddr.md), covering every hash size
// for the P2PKH and P2SH types across prefixes.
var cashAddrSpecVectors = []struct {
	addr   string
	prefix string
	p2sh   bool
	hash   string
}{
	{"bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2", "bitcoincash", false, "F5BF48B397DAE70BE82B3CCA4793F8EB2B6CDAC9"},
	{"bchtest:pr6m7j9njldwwzlg9v7v53unlr4jkmx6eyvwc0uz5t", "bchtest", true, "F5BF48B397DAE70BE82B3CCA4793F8EB2B6CDAC9"},
	{"pref:pr6m7j9njldwwzlg9v7v53unlr4jkmx6ey65nvtks5", "pref", true, "F5BF48B397DAE70BE82B3CCA4793F8EB2B6CDAC9"},
	{"bitcoincash:q9adhakpwzztepkpwp5z0dq62m6u5v5xtyj7j3h2ws4mr9g0", "bitcoincash", false, "7ADBF6C17084BC86C1706827B41A56F5CA32865925E946EA"},
	{"bchtest:p9adhakpwzztepkpwp5z0dq62m6u5v5xtyj7j3h2u94tsynr", "bchtest", true, "7ADBF6C17084BC86C1706827B41A56F5CA32865925E946EA"},
	{"bitcoincash:qgagf7w02x4wnz3mkwnchut2vxphjzccwxgjvvjmlsxqwkcw59jxxuz", "bitcoincash", false, "3A84F9CF51AAE98A3BB3A78BF16A6183790B18719126325BFC0C075B"},
	{"bchtest:pgagf7w02x4wnz3mkwnchut2vxphjzccwxgjvvjmlsxqwkcvs7md7wt", "bchtest", true, "3A84F9CF51AAE98A3BB3A78BF16A6183790B18719126325BFC0C075B"},
	{"bitcoincash:qvch8mmxy0rtfrlarg7ucrxxfzds5pamg73h7370aa87d80gyhqxq5nlegake", "bitcoincash", false, "3173EF6623C6B48FFD1A3DCC0CC6489B0A07BB47A37F47CFEF4FE69DE825C060"},
	{"bchtest:pvch8mmxy0rtfrlarg7ucrxxfzds5pamg73h7370aa87d80gyhqxq7fqng6m6", "bchtest", true, "3173EF6623C6B48FFD1A3DCC0CC6489B0A07BB47A37F47CFEF4FE69DE825C060"},
	{"pref:pvch8mmxy0rtfrlarg7ucrxxfzds5pamg73h7370aa87d80gyhqxq4k9m7qf9", "pref", true, "3173EF6623C6B48FFD1A3DCC0CC6489B0A07BB47A37F47CFEF4FE69DE825C060"},
	{"bitcoincash:qnq8zwpj8cq05n7pytfmskuk9r4gzzel8qtsvwz79zdskftrzxtar994cgutavfklv39gr3uvz", "bitcoincash", false, "C07138323E00FA4FC122D3B85B9628EA810B3F381706385E289B0B25631197D194B5C238BEB136FB"},
	{"bchtest:pnq8zwpj8cq05n7pytfmskuk9r4gzzel8qtsvwz79zdskftrzxtar994cgutavfklvmgm6ynej", "bchtest", true, "C07138323E00FA4FC122D3B85B9628EA810B3F381706385E289B0B25631197D194B5C238BEB136FB"},
	{"bitcoincash:qh3krj5607v3qlqh5c3wq3lrw3wnuxw0sp8dv0zugrrt5a3kj6ucysfz8kxwv2k53krr7n933jfsunqex2w82sl", "bitcoincash", false, "E361CA9A7F99107C17A622E047E3745D3E19CF804ED63C5C40C6BA763696B98241223D8CE62AD48D863F4CB18C930E4C"},
	{"bchtest:ph3krj5607v3qlqh5c3wq3lrw3wnuxw0sp8dv0zugrrt5a3kj6ucysfz8kxwv2k53krr7n933jfsunqnzf7mt6x", "bchtest", true, "E361CA9A7F99107C17A622E047E3745D3E19CF804ED63C5C40C6BA763696B98241223D8CE62AD48D863F4CB18C930E4C"},
	{"bitcoincash:qmvl5lzvdm6km38lgga64ek5jhdl7e3aqd9895wu04fvhlnare5937w4ywkq57juxsrhvw8ym5d8qx7sz7zz0zvcypqscw8jd03f", "bitcoincash", false, "D9FA7C4C6EF56DC4FF423BAAE6D495DBFF663D034A72D1DC7D52CBFE7D1E6858F9D523AC0A7A5C34077638E4DD1A701BD017842789982041"},
	{"bchtest:pmvl5lzvdm6km38lgga64ek5jhdl7e3aqd9895wu04fvhlnare5937w4ywkq57juxsrhvw8ym5d8qx7sz7zz0zvcypqs6kgdsg2g", "bchtest", true, "D9FA7C4C6EF56DC4FF423BAAE6D495DBFF663D034A72D1DC7D52CBFE7D1E6858F9D523AC0A7A5C34077638E4DD1A701BD017842789982041"},
	{"bitcoincash:qlg0x333p4238k0qrc5ej7rzfw5g8e4a4r6vvzyrcy8j3s5k0en7calvclhw46hudk5flttj6ydvjc0pv3nchp52amk97tqa5zygg96mtky5sv5w", "bitcoincash", false, "D0F346310D5513D9E01E299978624BA883E6BDA8F4C60883C10F28C2967E67EC77ECC7EEEAEAFC6DA89FAD72D11AC961E164678B868AEEEC5F2C1DA08884175B"},
	{"bchtest:plg0x333p4238k0qrc5ej7rzfw5g8e4a4r6vvzyrcy8j3s5k0en7calvclhw46hudk5flttj6ydvjc0pv3nchp52amk97tqa5zygg96mc773cwez", "bchtest", true, "D0F346310D5513D9E01E299978624BA883E6BDA8F4C60883C10F28C2967E67EC77ECC7EEEAEAFC6DA89FAD72D11AC961E164678B868AEEEC5F2C1DA08884175B"},
}

func TestCashAddrSpecVectors(t *testing.T) {
	for _, v := range cashAddrSpecVectors {
		wantHash, _ := hex.DecodeString(v.hash)
		prefix, p2sh, hash, err := decodeCashAddr(v.addr)
		if err != nil {
			t.Errorf("%s: decode error: %v", v.addr, err)
			continue
		}
		if prefix != v.prefix || p2sh != v.p2sh || !strings.EqualFold(hex.EncodeToString(hash), v.hash) {
			t.Errorf("%s: decoded prefix %s p2sh %t hash %x, want %s %t %x", v.addr, prefix, p2sh, hash, v.prefix, v.p2sh, wantHash)
			continue
		}
		if encoded := encodeCashAddr(v.prefix, v.p2sh, wantHash); encoded != v.addr {
			t.Errorf("%s: re-encoded as %s", v.addr, encoded)
		}
	}
}

// The spec vectors use "prefix:0..." for type 15, which is not a known type.
func TestCashAddrUnknownType(t *testing.T) {
	if _, _, _, err := decodeCashAddr("prefix:0r6m7j9njldwwzlg9v7v53unlr4jkmx6ey3qnjwsrf"); err == nil {
		t.Error("expected error for unknown address type")
	}
}

// Token-aware types (2 and 3) carry the same hash as their plain forms.
// bchutil cannot encode them, so the version byte is packed here.
func TestCashAddrTokenAware(t *testing.T) {
	hash, _ := hex.DecodeString("F5BF48B397DAE70BE82B3CCA4793F8EB2B6CDAC9")
	for _, test := range []struct {
		typeBits byte
		p2sh     bool
	}{{cashAddrTypeP2PKHToken, false}, {cashAddrTypeP2SHToken, true}} {
		payload, err := convertBits(append([]byte{test.typeBits << 3}, hash...), 8, 5, true)
		if err != nil {
			t.Fatalf("error packing token-aware payload: %v", err)
		}
		addr := bchutil.Encode("bitcoincash", payload)
		prefix, p2sh, gotHash, err := decodeCashAddr(addr)
		if err != nil {
			t.Errorf("%s: decode error: %v", addr, err)
			continue
		}
		if prefix != "bitcoincash" || p2sh != test.p2sh || hex.EncodeToString(gotHash) != strings.ToLower(hex.EncodeToString(hash)) {
			t.Errorf("%s: decoded prefix %s p2sh %t hash %x", addr, prefix, p2sh, gotHash)
		}
	}
}

func TestCashAddrPrefixOptional(t *testing.T) {
	with := "bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2"
	_, _, hashWith, err := decodeCashAddr(with)
	if err != nil {
		t.Fatalf("decode with prefix: %v", err)
	}
	_, _, hashWithout, err := decodeCashAddr(strings.TrimPrefix(with, "bitcoincash:"))
	if err != nil {
		t.Fatalf("decode without prefix: %v", err)
	}
	if hex.EncodeToString(hashWith) != hex.EncodeToString(hashWithout) {
		t.Errorf("hash differs with/without prefix: %x %x", hashWith, hashWithout)
	}
	if _, _, _, err := decodeCashAddr("bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg3"); err == nil {
		t.Error("expected checksum error")
	}
}
