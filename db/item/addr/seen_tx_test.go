package addr_test

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

var testTxHash = [32]byte{0xab, 0xcd, 0x01}

func legacyAddr(t *testing.T) wallet.LockAddr {
	t.Helper()
	address, err := wallet.GetAddrFromString("19Y1fAy7n2Qz1JhmRUWmWvDKuKrxL6u4J8")
	if err != nil {
		t.Fatalf("error parsing legacy address: %v", err)
	}
	return address.LockAddr()
}

func p2sh32Addr(t *testing.T) wallet.LockAddr {
	t.Helper()
	hash, _ := hex.DecodeString("0ab476a6dab00ba11118a11c078f871c5060a66e18140368425202772d0e80c2")
	addr32, err := wallet.GetAddr32FromScriptHash(hash)
	if err != nil {
		t.Fatalf("error building p2sh32 address: %v", err)
	}
	return addr32.LockAddr()
}

// zeroP2sh32Addr is the P2SH32 address of the all-zero script hash, which
// must keep its kind and 33-byte key like any other.
func zeroP2sh32Addr(t *testing.T) wallet.LockAddr {
	t.Helper()
	addr32, err := wallet.GetAddr32FromScriptHash(make([]byte, 32))
	if err != nil {
		t.Fatalf("error building zero hash p2sh32 address: %v", err)
	}
	return addr32.LockAddr()
}

// A legacy row's uid is unchanged from the fixed 25-byte layout (65 bytes),
// and a P2SH32 row's uid is 8 bytes longer; both read back exactly.
func TestSeenTxUid(t *testing.T) {
	seen := time.Unix(1700000000, 123456789)
	for _, test := range []struct {
		addr   wallet.LockAddr
		uidLen int
	}{{legacyAddr(t), 65}, {p2sh32Addr(t), 73}, {zeroP2sh32Addr(t), 73}} {
		seenTx := &addr.SeenTx{Addr: test.addr, Seen: seen, TxHash: testTxHash}
		uid := seenTx.GetUid()
		if len(uid) != test.uidLen {
			t.Errorf("%s: uid length %d, want %d", test.addr, len(uid), test.uidLen)
		}
		if !bytes.HasPrefix(uid, test.addr.Bytes()) {
			t.Errorf("%s: uid does not start with the address bytes", test.addr)
		}
		var parsed addr.SeenTx
		parsed.SetUid(uid)
		if parsed.Addr != test.addr || !parsed.Seen.Equal(seen) || parsed.TxHash != testTxHash {
			t.Errorf("%s: uid round trip mismatch: %+v", test.addr, parsed)
		}
		if parsed.GetShardSource() != seenTx.GetShardSource() {
			t.Errorf("%s: shard source changed through round trip", test.addr)
		}
	}
	var bad addr.SeenTx
	bad.SetUid(make([]byte, 70))
	if !bad.Addr.IsZero() {
		t.Error("expected uid of unknown length to be ignored")
	}
}

func TestP2shLinkRoundTrip(t *testing.T) {
	p2sh := wallet.GetAddrFromScriptHash(bytes.Repeat([]byte{0x11}, 20))
	addr32, _ := wallet.GetAddr32FromScriptHash(bytes.Repeat([]byte{0x22}, 32))
	links := addr.NewP2shLinks(*p2sh, *addr32)
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	for _, obj := range links {
		link := obj.(*addr.P2shLink)
		var parsed addr.P2shLink
		parsed.SetUid(link.GetUid())
		parsed.Deserialize(link.Serialize())
		if parsed.Addr != link.Addr || parsed.Other != link.Other {
			t.Errorf("link round trip mismatch: %+v vs %+v", parsed, *link)
		}
		if parsed.Addr == parsed.Other || parsed.Addr.IsP2SH32() == parsed.Other.IsP2SH32() {
			t.Errorf("link does not map between widths: %+v", parsed)
		}
	}
}
