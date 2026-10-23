package wallet_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/jchavannes/btcd/txscript"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Mainnet P2SH32 output 0 of tx
// 0000000000296205c93886e24d0964e96c24b5319e79300ee73b04f4ede74064 (block
// 974077), spent by input 0 of
// 00000000003ecac0815b55df13b88ee7b7c19470518a26a12dfd1004d77f85b9. The lock
// script here is the bare OP_HASH256 <32> OP_EQUAL part; on chain it carried a
// CashTokens prefix in front.
const (
	p2sh32HashHex   = "0ab476a6dab00ba11118a11c078f871c5060a66e18140368425202772d0e80c2"
	p2sh32LockHex   = "aa20" + p2sh32HashHex + "87"
	p2sh32UnlockHex = "210267d95a91e32a4cd512e20628979e3b8814ce70ef643a0bd0484ab68ed30fcbea004d1101c2529dc0cdc0c788c0d1c0ce88c351a063c0ccc0c602401f93a269c0d2c0cf88c0d3c0d09d6776b2757603ffff009f69c4529dc0ccc0c602dc0594a269c0d3c0d0c0d003a068069694a269c0d2827701649dc0d201247f78547f777853795679bb7681c0cf547f7701207f75816e5779028f009395029000969d51d3827754a06301ff51d358807e776701fe51d354807e7768c25480c37e00c87e00c954807e01fd7e00ca8277527f757e00ca7e00cb54807ec47e00cc58807e02b3ef7e00d101207f757e01717e01647e00d27e01ff7e00d358807e00cd7e51cc58807e012251cd82779352798277937e01ef7e51d17e607e787e51cd7e0054807e76aa01207f75817600a0697777789f696d6d6d6851"
)

var p2sh32Lock, _ = hex.DecodeString(p2sh32LockHex)
var p2sh32Unlock, _ = hex.DecodeString(p2sh32UnlockHex)
var p2sh32Hash, _ = hex.DecodeString(p2sh32HashHex)

func TestP2sh32LockScript(t *testing.T) {
	addr32, err := wallet.GetP2sh32AddrFromLockScript(p2sh32Lock)
	if err != nil {
		t.Fatalf("error getting p2sh32 addr from lock script: %v", err)
	}
	if !bytes.Equal(addr32[:], p2sh32Hash) {
		t.Fatalf("p2sh32 hash mismatch: %x", addr32[:])
	}
	lockAddr, err := wallet.GetLockAddrFromLockScript(p2sh32Lock)
	if err != nil {
		t.Fatalf("error getting lock addr from p2sh32 lock script: %v", err)
	}
	if !lockAddr.IsP2SH32() || lockAddr.Addr32() != *addr32 {
		t.Fatalf("lock addr mismatch: %+v", lockAddr)
	}
	// Round trip through the string form (CashAddr, 256-bit P2SH) and bytes.
	parsed, err := wallet.GetLockAddrFromString(lockAddr.String())
	if err != nil {
		t.Fatalf("error parsing %s: %v", lockAddr.String(), err)
	}
	if parsed != lockAddr {
		t.Fatalf("string round trip mismatch: %s -> %+v", lockAddr.String(), parsed)
	}
	fromBytes, err := wallet.GetLockAddrFromBytes(lockAddr.Bytes())
	if err != nil || fromBytes != lockAddr {
		t.Fatalf("bytes round trip mismatch: %x -> %+v %v", lockAddr.Bytes(), fromBytes, err)
	}
	if b := lockAddr.Bytes(); len(b) != wallet.LockAddrLengthP2sh32 || b[0] != wallet.AddrVersionP2SH32 {
		t.Fatalf("unexpected p2sh32 key bytes: %x", b)
	}
	t.Logf("p2sh32 address: %s", lockAddr)
}

// The spend's redeem script hashes (HASH256) to the output's committed hash,
// and its HASH160 is the P2SH20 alias the index records inputs under.
func TestP2sh32UnlockScript(t *testing.T) {
	redeemScript, err := wallet.GetP2shRedeemScriptFromUnlockScript(p2sh32Unlock)
	if err != nil {
		t.Fatalf("error getting redeem script: %v", err)
	}
	addr, addr32 := wallet.GetP2shAddrsFromRedeemScript(redeemScript)
	if !bytes.Equal(addr32[:], p2sh32Hash) {
		t.Fatalf("redeem script HASH256 %x does not match lock hash %x", addr32[:], p2sh32Hash)
	}
	if !addr.IsP2SH() {
		t.Fatalf("expected p2sh20 alias, got %s", addr)
	}
	// The legacy unlock path yields the same alias.
	aliasAddr, err := wallet.GetAddrFromUnlockScript(p2sh32Unlock)
	if err != nil {
		t.Fatalf("error getting addr from unlock script: %v", err)
	}
	if *aliasAddr != addr {
		t.Fatalf("alias mismatch: %s vs %s", aliasAddr, addr)
	}
	t.Logf("p2sh20 alias: %s, p2sh32: %s", addr, addr32)
}

func TestLockAddrLegacy(t *testing.T) {
	for _, test := range tests {
		lockAddr, err := wallet.GetLockAddrFromLockScript(test.LockScript)
		if err != nil {
			t.Errorf("error getting lock addr for %s: %v", test.Address, err)
			continue
		}
		if lockAddr.IsP2SH32() || lockAddr.String() != test.Address {
			t.Errorf("lock addr mismatch: %s vs %s", lockAddr, test.Address)
		}
		// Key bytes are exactly the legacy 25 bytes, so existing rows match.
		legacy := lockAddr.Addr()
		if !bytes.Equal(lockAddr.Bytes(), legacy[:]) || len(lockAddr.Bytes()) != wallet.LockAddrLengthLegacy {
			t.Errorf("legacy key bytes changed: %x", lockAddr.Bytes())
		}
		parsed, err := wallet.GetLockAddrFromString(test.Address)
		if err != nil || parsed != lockAddr {
			t.Errorf("base58 parse mismatch for %s: %+v %v", test.Address, parsed, err)
		}
		// The same address as a 20-byte CashAddr resolves to the same Addr.
		cashAddr := lockAddr.Addr().OldAddress().GetCashAddrString()
		parsed, err = wallet.GetLockAddrFromString(cashAddr)
		if err != nil || parsed != lockAddr {
			t.Errorf("cash addr parse mismatch for %s: %+v %v", cashAddr, parsed, err)
		}
	}
	if _, err := wallet.GetLockAddrFromLockScript([]byte{txscript.OP_RETURN, 0x01, 0x00}); err == nil {
		t.Error("expected error for op return lock script")
	}
	if _, err := wallet.GetLockAddrFromString("bchtest:pr6m7j9njldwwzlg9v7v53unlr4jkmx6eyvwc0uz5t"); err == nil {
		t.Error("expected error for testnet prefix")
	}
	if _, err := wallet.GetLockAddrFromString("bitcoincash:qvch8mmxy0rtfrlarg7ucrxxfzds5pamg73h7370aa87d80gyhqxq5nlegake"); err == nil {
		t.Error("expected error for 32-byte p2pkh address")
	}
	if _, err := wallet.GetLockAddrFromBytes(make([]byte, 33)); err == nil {
		t.Error("expected error for 33 bytes without p2sh32 version")
	}
	var zero wallet.LockAddr
	if !zero.IsZero() || zero.IsP2SH32() {
		t.Error("zero lock addr misreported")
	}
}

func BenchmarkAddr32String(b *testing.B) {
	addr32, _ := wallet.GetAddr32FromScriptHash(p2sh32Hash)
	for i := 0; i < b.N; i++ {
		_ = addr32.String()
	}
}
