package wallet

import (
	"fmt"

	"github.com/jchavannes/btcutil/base58"
)

// LockAddr is the address a lock script pays to, in either of the index's
// two widths: a legacy Addr (P2PKH or P2SH20) or an Addr32 (P2SH32). It is
// the key form itself, padded to the wider kind: the first byte is the
// version and says which layout follows, so the kind never depends on a
// length.
//
//	legacy: version (0 or 5), 20-byte hash, 4-byte checksum, 8 zero bytes
//	p2sh32: AddrVersionP2SH32, 32-byte hash
//
// Bytes returns the unpadded key, which for a legacy address is its 25
// bytes exactly as stored before P2SH32 existed. The zero value is no
// address (as for Addr), while the P2SH32 address of an all-zero hash keeps
// its version byte and so stays distinct. It is comparable so it can key
// maps, and it is the type stored where a topic key may be either width.
//
// The memo topics never hold a P2SH32 address: memo actions are attributed
// from unlock scripts, which only ever yield the 20-byte form (a P2SH32
// spend is recorded under the P2SH20 alias of its redeem script, see
// addr.P2shLink), so those items keep plain Addr fields.
type LockAddr [LockAddrLengthP2sh32]byte

const (
	LockAddrLengthLegacy = 25
	LockAddrLengthP2sh32 = 33
)

// LockAddrLength is the key length a version byte announces: 33 for
// AddrVersionP2SH32, 25 for any legacy version. Uids that embed an address
// read this from their first byte rather than inferring the kind from a
// length.
func LockAddrLength(version byte) int {
	if version == AddrVersionP2SH32 {
		return LockAddrLengthP2sh32
	}
	return LockAddrLengthLegacy
}

// LockAddr wraps the address in the union type used where either width is
// accepted.
func (a Addr) LockAddr() LockAddr {
	var l LockAddr
	copy(l[:], a[:])
	return l
}

// LockAddr wraps the address in the union type used where either width is
// accepted.
func (a Addr32) LockAddr() LockAddr {
	var l LockAddr
	l[0] = AddrVersionP2SH32
	copy(l[1:], a[:])
	return l
}

func (l LockAddr) IsP2SH32() bool {
	return l[0] == AddrVersionP2SH32
}

// IsP2SH reports whether the address is a script hash of either width.
func (l LockAddr) IsP2SH() bool {
	return l.IsP2SH32() || l.Addr().IsP2SH()
}

func (l LockAddr) IsZero() bool {
	return l == LockAddr{}
}

// Addr is the legacy address; zero for a P2SH32 address.
func (l LockAddr) Addr() Addr {
	var a Addr
	if !l.IsP2SH32() {
		copy(a[:], l[:])
	}
	return a
}

// Addr32 is the P2SH32 address; zero for a legacy address.
func (l LockAddr) Addr32() Addr32 {
	var a Addr32
	if l.IsP2SH32() {
		copy(a[:], l[1:])
	}
	return a
}

func (l LockAddr) String() string {
	if l.IsP2SH32() {
		return l.Addr32().String()
	}
	return l.Addr().String()
}

// Bytes is the key form: the unchanged 25 legacy bytes for a 20-byte
// address, or AddrVersionP2SH32 followed by the 32-byte hash.
func (l LockAddr) Bytes() []byte {
	return l[:LockAddrLength(l[0])]
}

// GetLockAddrFromBytes parses key bytes written by Bytes. The version byte
// decides the kind, and the length must be exactly what it announces.
func GetLockAddrFromBytes(b []byte) (LockAddr, error) {
	if len(b) == 0 {
		return LockAddr{}, fmt.Errorf("error lock addr bytes are empty")
	}
	if size := LockAddrLength(b[0]); len(b) != size {
		return LockAddr{}, fmt.Errorf("error lock addr bytes have length %d, version %d expects %d", len(b), b[0], size)
	}
	var l LockAddr
	copy(l[:], b)
	return l, nil
}

// GetLockAddrFromString accepts a legacy base58 address or a CashAddr (with
// or without the bitcoincash: prefix, token-aware or not). A 20-byte
// CashAddr resolves to the same Addr as its base58 form; only a 32-byte P2SH
// CashAddr yields an Addr32.
func GetLockAddrFromString(addrString string) (LockAddr, error) {
	if d := base58.Decode(addrString); len(d) == LockAddrLengthLegacy {
		var addr Addr
		copy(addr[:], d)
		return addr.LockAddr(), nil
	}
	prefix, p2sh, hash, err := decodeCashAddr(addrString)
	if err != nil {
		return LockAddr{}, fmt.Errorf("error address is neither base58 nor cash address; %w", err)
	}
	if prefix != CashAddrPrefix {
		return LockAddr{}, fmt.Errorf("error cash address prefix is not %s: %s", CashAddrPrefix, prefix)
	}
	switch {
	case len(hash) == 20 && p2sh:
		return GetAddrFromScriptHash(hash).LockAddr(), nil
	case len(hash) == 20:
		return GetAddrFromPkHash(hash).LockAddr(), nil
	case len(hash) == 32 && p2sh:
		addr32, err := GetAddr32FromScriptHash(hash)
		if err != nil {
			return LockAddr{}, err
		}
		return addr32.LockAddr(), nil
	}
	return LockAddr{}, fmt.Errorf("error unsupported cash address: p2sh %t, hash size %d", p2sh, len(hash))
}

// GetLockAddrFromLockScript resolves a lock script of any supported kind:
// P2PKH, P2SH20 or P2SH32.
func GetLockAddrFromLockScript(lockScript []byte) (LockAddr, error) {
	if addr, err := GetAddrFromLockScript(lockScript); err == nil {
		return addr.LockAddr(), nil
	}
	addr32, err := GetP2sh32AddrFromLockScript(lockScript)
	if err != nil {
		return LockAddr{}, fmt.Errorf("error lock script is not a p2pkh, p2sh or p2sh32 address; %w", err)
	}
	return addr32.LockAddr(), nil
}

// LockAddrsToBytes returns each address's key bytes, e.g. for prefix scans.
func LockAddrsToBytes(addrs []LockAddr) [][]byte {
	var addrBytes = make([][]byte, len(addrs))
	for i := range addrs {
		addrBytes[i] = addrs[i].Bytes()
	}
	return addrBytes
}
