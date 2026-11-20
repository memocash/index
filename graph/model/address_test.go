package model_test

import (
	"testing"

	"github.com/memocash/index/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// The scalar keeps a P2SH32 address's kind through unmarshal and marshal,
// including the all-zero script hash, whose payload alone looks like no
// address.
func TestAddressScalarRoundTrip(t *testing.T) {
	zero32, _ := wallet.GetAddr32FromScriptHash(make([]byte, 32))
	legacy, _ := wallet.GetAddrFromString("19Y1fAy7n2Qz1JhmRUWmWvDKuKrxL6u4J8")
	for _, want := range []wallet.LockAddr{zero32.LockAddr(), legacy.LockAddr()} {
		parsed, err := model.UnmarshalAddress(want.String())
		if err != nil {
			t.Fatalf("error unmarshalling %s; %v", want, err)
		}
		if parsed.LockAddr() != want || parsed.IsZero() || parsed.IsP2SH32() != want.IsP2SH32() {
			t.Fatalf("%s unmarshalled to %+v", want, parsed)
		}
		if parsed.String() != want.String() {
			t.Fatalf("%s marshals back as %s", want, parsed.String())
		}
	}
}
