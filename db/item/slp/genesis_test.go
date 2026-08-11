package slp_test

import (
	"github.com/memocash/index/db/item/slp"
	"testing"
)

func TestGenesisSerializeRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		genesis slp.Genesis
	}{{
		name: "all fields",
		genesis: slp.Genesis{
			TxHash:     [32]byte{0xab, 0xcd, 0xef, 0x01},
			TokenType:  0x81,
			Decimals:   8,
			BatonIndex: 2,
			Ticker:     "TEST",
			Name:       "Test Token",
			DocUrl:     "https://example.com",
			DocHash:    [32]byte{0x01, 0x02},
		},
	}, {
		name: "empty variable fields",
		genesis: slp.Genesis{
			TokenType: 0x01,
			Decimals:  0,
		},
	}, {
		name: "null bytes preserved",
		genesis: slp.Genesis{
			TokenType: 0x01,
			Ticker:    string([]byte{'A', 0x00, 'B'}),
			Name:      string([]byte{0x00, 0x00}),
			DocUrl:    string([]byte{0x00, 'x'}),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var read slp.Genesis
			read.SetUid(tt.genesis.GetUid())
			read.Deserialize(tt.genesis.Serialize())
			if read != tt.genesis {
				t.Errorf("round trip mismatch: got %+v, want %+v", read, tt.genesis)
			}
		})
	}
}

func TestGenesisDeserializeMalformed(t *testing.T) {
	valid := (&slp.Genesis{Ticker: "TEST", Name: "Test"}).Serialize()
	for _, tt := range []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "short fixed prefix", data: valid[:10]},
		{name: "truncated field", data: valid[:len(valid)-1]},
		{name: "trailing garbage", data: append(append([]byte{}, valid...), 0xff)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var read slp.Genesis
			read.Deserialize(tt.data)
			if read != (slp.Genesis{}) {
				t.Errorf("malformed data should leave genesis zero, got %+v", read)
			}
		})
	}
}
