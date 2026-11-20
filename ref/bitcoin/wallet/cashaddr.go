package wallet

import (
	"fmt"
	"strings"

	"github.com/jchavannes/bchutil"
)

// cashAddrHashSizes maps the three size bits of a CashAddr version byte to
// the hash length in bytes.
var cashAddrHashSizes = [8]int{20, 24, 28, 32, 40, 48, 56, 64}

// CashAddr version byte type bits (bits 3-6). Types 2 and 3 are the
// token-aware forms introduced with CashTokens; they carry the same hash as
// types 0 and 1 and only signal that the receiver accepts tokens.
const (
	cashAddrTypeP2PKH      = 0
	cashAddrTypeP2SH       = 1
	cashAddrTypeP2PKHToken = 2
	cashAddrTypeP2SHToken  = 3
)

// encodeCashAddr returns the CashAddr string (with prefix) for a P2PKH or
// P2SH hash of any size the format allows, or "" for a size it does not.
// bchutil's encoder mis-sizes hashes over 32 bytes and builds the string a
// character at a time, so the version byte is packed and the string written
// here; only the checksum step is shared. The checksum is a 40-bit BCH code
// over the prefix and payload (not a hash), cheap enough to compute on each
// call rather than store with the address.
func encodeCashAddr(prefix string, p2sh bool, hash []byte) string {
	var version byte
	if p2sh {
		version = cashAddrTypeP2SH << 3
	}
	var sizeOk bool
	for sizeBits, size := range cashAddrHashSizes {
		if size == len(hash) {
			version |= byte(sizeBits)
			sizeOk = true
			break
		}
	}
	if !sizeOk {
		return ""
	}
	payload, err := convertBits(append([]byte{version}, hash...), 8, 5, true)
	if err != nil {
		return ""
	}
	checksum := bchutil.CreateChecksum(prefix, payload)
	var out = make([]byte, 0, len(prefix)+1+len(payload)+len(checksum))
	out = append(out, prefix...)
	out = append(out, ':')
	for _, c := range payload {
		out = append(out, bchutil.CHARSET[c])
	}
	for _, c := range checksum {
		out = append(out, bchutil.CHARSET[c])
	}
	return string(out)
}

// decodeCashAddr parses a CashAddr string into its prefix, script type and
// hash. The prefix may be omitted, in which case the mainnet prefix is
// assumed. Token-aware types resolve to the plain type with the same hash.
// bchutil only decodes 20-byte hashes, so the version byte is unpacked here.
func decodeCashAddr(addrString string) (prefix string, p2sh bool, hash []byte, err error) {
	if !strings.Contains(addrString, ":") {
		addrString = CashAddrPrefix + ":" + addrString
	}
	prefix, values, err := bchutil.DecodeCashAddress(addrString)
	if err != nil {
		return "", false, nil, fmt.Errorf("error decoding cash address; %w", err)
	}
	payload, err := convertBits(values, 5, 8, false)
	if err != nil {
		return "", false, nil, fmt.Errorf("error converting cash address bits; %w", err)
	}
	if len(payload) == 0 {
		return "", false, nil, fmt.Errorf("error cash address has no payload")
	}
	version := payload[0]
	if version&0x80 != 0 {
		return "", false, nil, fmt.Errorf("error cash address version byte has reserved bit set")
	}
	hash = payload[1:]
	if size := cashAddrHashSizes[version&0x07]; len(hash) != size {
		return "", false, nil, fmt.Errorf("error cash address hash size %d does not match version size %d", len(hash), size)
	}
	switch (version >> 3) & 0x0f {
	case cashAddrTypeP2PKH, cashAddrTypeP2PKHToken:
		p2sh = false
	case cashAddrTypeP2SH, cashAddrTypeP2SHToken:
		p2sh = true
	default:
		return "", false, nil, fmt.Errorf("error cash address has unknown type %d", (version>>3)&0x0f)
	}
	return prefix, p2sh, hash, nil
}

// convertBits regroups a byte slice of fromBits-wide values into toBits-wide
// values, as in the bech32 reference implementation. Without padding, any
// non-zero leftover bits are an error.
func convertBits(data []byte, fromBits, toBits uint, pad bool) ([]byte, error) {
	var acc, bits uint
	var ret []byte
	maxV := uint(1<<toBits) - 1
	for _, value := range data {
		if uint(value)>>fromBits != 0 {
			return nil, fmt.Errorf("error invalid value %d for %d bits", value, fromBits)
		}
		acc = acc<<fromBits | uint(value)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			ret = append(ret, byte(acc>>bits&maxV))
		}
	}
	if pad {
		if bits > 0 {
			ret = append(ret, byte(acc<<(toBits-bits)&maxV))
		}
	} else if bits >= fromBits || acc<<(toBits-bits)&maxV != 0 {
		return nil, fmt.Errorf("error invalid padding in bit conversion")
	}
	return ret, nil
}
