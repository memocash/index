package cashtoken_test

import (
	"bytes"
	"testing"

	item "github.com/memocash/index/db/item/cashtoken"
	"github.com/memocash/index/ref/bitcoin/tx/cashtoken"
)

func TestOutputSerializeRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		token cashtoken.Token
	}{
		{"fungible", cashtoken.Token{Category: [32]byte{1, 2, 3}, Amount: 1000}},
		{"immutable nft", cashtoken.Token{Category: [32]byte{9}, HasNft: true}},
		{"minting nft with commitment and amount", cashtoken.Token{
			Category: [32]byte{0xff}, HasNft: true, Capability: cashtoken.CapabilityMinting,
			Commitment: []byte{0xab, 0xcd, 0xef}, Amount: 7}},
		{"mutable nft single byte commitment", cashtoken.Token{
			HasNft: true, Capability: cashtoken.CapabilityMutable, Commitment: []byte{0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := item.NewOutput([32]byte{0xaa, 0xbb}, 3, tt.token)
			var got item.Output
			got.SetUid(want.GetUid())
			got.Deserialize(want.Serialize())
			if got.TxHash != want.TxHash || got.Index != want.Index || got.Category != want.Category ||
				got.HasNft != want.HasNft || got.Capability != want.Capability || got.Amount != want.Amount ||
				!bytes.Equal(got.Commitment, want.Commitment) {
				t.Errorf("round trip mismatch: got %+v, want %+v", got, *want)
			}
			if !bytes.Equal(cashtoken.Encode(got.Token()), cashtoken.Encode(tt.token)) {
				t.Errorf("token prefix mismatch after round trip")
			}
		})
	}
}

func TestOutputDeserializeShort(t *testing.T) {
	var out item.Output
	out.Deserialize(make([]byte, 40))
	if out.Amount != 0 || out.HasNft || out.Commitment != nil {
		t.Errorf("short data should leave output zero: %+v", out)
	}
}
