package cashtoken_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/memocash/index/ref/bitcoin/tx/cashtoken"
)

const (
	category = "0000000000000000000000000000000000000000000000000000000000000001"
	p2pkh    = "76a91489abcdefabbaabbaabbaabbaabbaabbaabbaabba88ac"
)

func fromHex(t *testing.T, s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

var categoryBytes = func() [32]byte {
	var c [32]byte
	c[31] = 1
	return c
}()

// Prefix vectors mirror btcd's txscript/cashtoken_test.go so the two
// decoders agree on what is well formed.
func TestParseValid(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   cashtoken.Token
	}{
		{"fungible amount 1", "10" + "01", cashtoken.Token{Amount: 1}},
		{"fungible amount 0xfd encoded", "10" + "fde803", cashtoken.Token{Amount: 1000}},
		{"fungible amount 0xfe encoded", "10" + "fe00e1f505", cashtoken.Token{Amount: 100000000}},
		{"fungible max amount", "10" + "ffffffffffffffff7f", cashtoken.Token{Amount: cashtoken.MaxAmount}},
		{"immutable nft no commitment", "20", cashtoken.Token{HasNft: true}},
		{"mutable nft with commitment", "61" + "02" + "abcd", cashtoken.Token{
			HasNft: true, Capability: cashtoken.CapabilityMutable, Commitment: []byte{0xab, 0xcd}}},
		{"minting nft with commitment and amount", "72" + "03" + "010203" + "fde803", cashtoken.Token{
			HasNft: true, Capability: cashtoken.CapabilityMinting, Commitment: []byte{1, 2, 3}, Amount: 1000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Category = categoryBytes
			script := fromHex(t, "ef"+category+tt.prefix+p2pkh)
			token, locking, err := cashtoken.Parse(script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token.Category != tt.want.Category || token.HasNft != tt.want.HasNft ||
				token.Capability != tt.want.Capability || token.Amount != tt.want.Amount ||
				!bytes.Equal(token.Commitment, tt.want.Commitment) {
				t.Errorf("token mismatch: got %+v, want %+v", *token, tt.want)
			}
			if !bytes.Equal(locking, fromHex(t, p2pkh)) {
				t.Errorf("locking bytecode mismatch: %x", locking)
			}
			if got := cashtoken.StripPrefix(script); !bytes.Equal(got, fromHex(t, p2pkh)) {
				t.Errorf("StripPrefix mismatch: %x", got)
			}
			if encoded := cashtoken.Encode(*token); !bytes.Equal(encoded, fromHex(t, "ef"+category+tt.prefix)) {
				t.Errorf("Encode mismatch: got %x, want %s", encoded, "ef"+category+tt.prefix)
			}
		})
	}
}

func TestParseMalformed(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"prefix only", "ef"},
		{"truncated category", "ef" + category[:20]},
		{"missing bitfield", "ef" + category},
		{"reserved bit set", "ef" + category + "90" + "01"},
		{"neither nft nor amount", "ef" + category + "00"},
		{"capability without nft", "ef" + category + "11" + "01"},
		{"commitment without nft", "ef" + category + "50" + "01" + "aa" + "01"},
		{"capability out of range", "ef" + category + "23"},
		{"missing commitment length", "ef" + category + "60"},
		{"zero commitment length", "ef" + category + "60" + "00"},
		{"commitment longer than script", "ef" + category + "60" + "05" + "0102"},
		{"commitment length past end", "ef" + category + "60" + "fd"},
		{"non-minimal commitment length", "ef" + category + "60" + "fd0100" + "aa"},
		{"missing amount", "ef" + category + "10"},
		{"zero amount", "ef" + category + "10" + "00"},
		{"amount past end", "ef" + category + "10" + "ff01020304"},
		{"non-minimal amount", "ef" + category + "10" + "fd0100"},
		{"amount above max", "ef" + category + "10" + "ff0000000000000080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := fromHex(t, tt.script)
			token, _, err := cashtoken.Parse(script)
			if !errors.Is(err, cashtoken.ErrMalformed) {
				t.Fatalf("expected ErrMalformed, got token %+v err %v", token, err)
			}
			if got := cashtoken.StripPrefix(script); !bytes.Equal(got, script) {
				t.Errorf("StripPrefix should leave malformed script unchanged, got %x", got)
			}
		})
	}
}

func TestParseNoPrefix(t *testing.T) {
	script := fromHex(t, p2pkh)
	token, locking, err := cashtoken.Parse(script)
	if err != nil || token != nil || !bytes.Equal(locking, script) {
		t.Errorf("plain script: token %v locking %x err %v", token, locking, err)
	}
	if cashtoken.HasPrefix(script) || cashtoken.HasPrefix(nil) {
		t.Error("HasPrefix false positive")
	}
	// 0xef later in a script is an ordinary (invalid) opcode, not a prefix.
	if _, _, err := cashtoken.Parse(append(fromHex(t, p2pkh), 0xef)); err != nil {
		t.Errorf("trailing 0xef should not parse as prefix: %v", err)
	}
}
