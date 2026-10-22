// Package cashtoken decodes the CashTokens output prefix
// (CHIP-2022-02-CashTokens). A token-carrying output's locking bytecode is
// preceded by:
//
//	OP_PREFIXTOKEN (0xef) <category (32 bytes)> <bitfield (1 byte)>
//	[<commitment length (CompactSize)> <commitment>] [<amount (CompactSize)>]
//
// The shared btcd fork recognises the prefix when parsing scripts (so address
// extraction works) but exposes none of the token fields, hence this decoder.
// Validation mirrors btcd's parseTokenPrefix so the index never records a
// token the node would have rejected.
package cashtoken

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
)

const (
	bitfieldReserved      = 0x80
	bitfieldHasCommitment = 0x40
	bitfieldHasNft        = 0x20
	bitfieldHasAmount     = 0x10
	bitfieldCapability    = 0x0f

	CategorySize = 32
	MaxAmount    = 1<<63 - 1
)

// Capability is an NFT's capability. Fungible-only tokens have none.
type Capability uint8

const (
	CapabilityNone    Capability = 0x00
	CapabilityMutable Capability = 0x01
	CapabilityMinting Capability = 0x02
)

func (c Capability) String() string {
	switch c {
	case CapabilityNone:
		return "NONE"
	case CapabilityMutable:
		return "MUTABLE"
	case CapabilityMinting:
		return "MINTING"
	}
	return fmt.Sprintf("UNKNOWN(%d)", uint8(c))
}

// Token is the decoded prefix of a token-carrying output. Category is the
// genesis tx hash in internal (little-endian) byte order, the same order the
// prefix carries it in. Amount is the fungible amount (0 when the output
// carries only an NFT). HasNft is false for fungible-only outputs; for NFTs
// Commitment is nil when the prefix has no commitment.
type Token struct {
	Category   [CategorySize]byte
	HasNft     bool
	Capability Capability
	Commitment []byte
	Amount     uint64
}

// Bitfield re-encodes the token's flag byte.
func (t Token) Bitfield() byte {
	var b byte
	if t.HasNft {
		b |= bitfieldHasNft | byte(t.Capability)
	}
	if len(t.Commitment) > 0 {
		b |= bitfieldHasCommitment
	}
	if t.Amount > 0 {
		b |= bitfieldHasAmount
	}
	return b
}

// HasPrefix reports whether a lock script starts with the token prefix byte.
// It does not check that the prefix is well formed; see Parse.
func HasPrefix(pkScript []byte) bool {
	return len(pkScript) > 0 && pkScript[0] == txscript.OP_PREFIXTOKEN
}

// ErrMalformed wraps every parse failure of a script that starts with the
// token prefix byte.
var ErrMalformed = errors.New("malformed cash token prefix")

func malformed(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrMalformed, fmt.Sprintf(format, args...))
}

// Parse splits a lock script into its token prefix and the locking bytecode
// that follows. A script without the prefix byte returns a nil token and the
// script unchanged. A script with the prefix byte but a malformed prefix
// returns an error wrapping ErrMalformed; such an output is unspendable
// and carries no token.
func Parse(pkScript []byte) (*Token, []byte, error) {
	if !HasPrefix(pkScript) {
		return nil, pkScript, nil
	}
	data := pkScript[1:]
	if len(data) < CategorySize+1 {
		return nil, nil, malformed("requires %d bytes, script has %d", CategorySize+1, len(data))
	}
	var token = new(Token)
	copy(token.Category[:], data[:CategorySize])
	bitfield := data[CategorySize]
	hasCommitment := bitfield&bitfieldHasCommitment != 0
	hasAmount := bitfield&bitfieldHasAmount != 0
	token.HasNft = bitfield&bitfieldHasNft != 0
	token.Capability = Capability(bitfield & bitfieldCapability)
	switch {
	case bitfield&bitfieldReserved != 0:
		return nil, nil, malformed("reserved bitfield bit set")
	case !token.HasNft && !hasAmount:
		return nil, nil, malformed("neither nft nor amount")
	case !token.HasNft && (hasCommitment || token.Capability != CapabilityNone):
		return nil, nil, malformed("commitment or capability without nft")
	case token.Capability > CapabilityMinting:
		return nil, nil, malformed("invalid capability %d", token.Capability)
	}
	n := CategorySize + 1
	readCompactSize := func() (uint64, error) {
		v, err := wire.ReadVarInt(bytes.NewReader(data[n:]), 0)
		if err != nil {
			return 0, err
		}
		n += wire.VarIntSerializeSize(v)
		return v, nil
	}
	if hasCommitment {
		commitmentLen, err := readCompactSize()
		if err != nil {
			return nil, nil, malformed("invalid commitment length: %v", err)
		}
		if commitmentLen == 0 {
			return nil, nil, malformed("commitment length must be at least 1")
		}
		if commitmentLen > uint64(len(data)-n) {
			return nil, nil, malformed("commitment requires %d bytes, script has %d", commitmentLen, len(data)-n)
		}
		token.Commitment = append([]byte{}, data[n:n+int(commitmentLen)]...)
		n += int(commitmentLen)
	}
	if hasAmount {
		amount, err := readCompactSize()
		if err != nil {
			return nil, nil, malformed("invalid amount: %v", err)
		}
		if amount == 0 || amount > MaxAmount {
			return nil, nil, malformed("amount %d out of range", amount)
		}
		token.Amount = amount
	}
	return token, data[n:], nil
}

// StripPrefix returns the locking bytecode of a lock script, dropping a
// well-formed token prefix if present. A script with a malformed prefix is
// returned unchanged, so callers parsing it as a plain script fail as before.
func StripPrefix(pkScript []byte) []byte {
	_, locking, err := Parse(pkScript)
	if err != nil {
		return pkScript
	}
	return locking
}

// Encode builds the token prefix for a token; appending the locking
// bytecode yields a complete lock script.
func Encode(token Token) []byte {
	var buf bytes.Buffer
	buf.WriteByte(txscript.OP_PREFIXTOKEN)
	buf.Write(token.Category[:])
	buf.WriteByte(token.Bitfield())
	if len(token.Commitment) > 0 {
		_ = wire.WriteVarInt(&buf, 0, uint64(len(token.Commitment)))
		buf.Write(token.Commitment)
	}
	if token.Amount > 0 {
		_ = wire.WriteVarInt(&buf, 0, token.Amount)
	}
	return buf.Bytes()
}
