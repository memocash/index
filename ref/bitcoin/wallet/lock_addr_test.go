package wallet_test

import (
	"bytes"
	"testing"

	"github.com/memocash/index/ref/bitcoin/wallet"
)

// An all-zero script hash is a valid P2SH32 address (anyone can pay to it,
// nobody can spend it). It must keep its kind through every form of the
// union: the payload alone cannot tell it from no address at all.
func TestLockAddrZeroHashP2sh32(t *testing.T) {
	var zeroHash = make([]byte, 32)
	addr32, err := wallet.GetAddr32FromScriptHash(zeroHash)
	if err != nil {
		t.Fatalf("error getting zero hash p2sh32 address; %v", err)
	}
	lockAddr := addr32.LockAddr()
	if !lockAddr.IsP2SH32() || lockAddr.IsZero() {
		t.Fatalf("zero hash p2sh32 lost its kind: p2sh32=%t zero=%t", lockAddr.IsP2SH32(), lockAddr.IsZero())
	}
	wantKey := append([]byte{wallet.AddrVersionP2SH32}, zeroHash...)
	if !bytes.Equal(lockAddr.Bytes(), wantKey) {
		t.Fatalf("zero hash p2sh32 key %x, want %x", lockAddr.Bytes(), wantKey)
	}
	fromBytes, err := wallet.GetLockAddrFromBytes(lockAddr.Bytes())
	if err != nil || fromBytes != lockAddr {
		t.Fatalf("zero hash p2sh32 bytes round trip: %+v %v", fromBytes, err)
	}
	cashAddr := lockAddr.String()
	if cashAddr != addr32.String() {
		t.Fatalf("zero hash p2sh32 string %s, want %s", cashAddr, addr32.String())
	}
	fromString, err := wallet.GetLockAddrFromString(cashAddr)
	if err != nil || fromString != lockAddr {
		t.Fatalf("zero hash p2sh32 string round trip of %s: %+v %v", cashAddr, fromString, err)
	}
	lockScript := append(append([]byte{0xaa, 0x20}, zeroHash...), 0x87)
	fromScript, err := wallet.GetLockAddrFromLockScript(lockScript)
	if err != nil || fromScript != lockAddr {
		t.Fatalf("zero hash p2sh32 lock script: %+v %v", fromScript, err)
	}
	// The zero value is still no address, and a legacy address is not p2sh32.
	if !(wallet.LockAddr{}).IsZero() || (wallet.LockAddr{}).IsP2SH32() {
		t.Fatalf("zero lock addr is not zero or claims p2sh32")
	}
	var zeroAddr wallet.Addr
	if zeroAddr.LockAddr().IsP2SH32() || !bytes.Equal(zeroAddr.LockAddr().Bytes(), make([]byte, 25)) {
		t.Fatalf("legacy zero address misclassified")
	}
}
