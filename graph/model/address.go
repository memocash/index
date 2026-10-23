package model

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

// Address is the GraphQL Address scalar: a P2PKH or P2SH20 address written
// as legacy base58, or a P2SH32 address written as a CashAddr (legacy base58
// has no P2SH32 form). Input accepts base58 or CashAddr for any kind.
type Address wallet.LockAddr

func (a Address) String() string {
	return wallet.LockAddr(a).String()
}

func (a Address) LockAddr() wallet.LockAddr {
	return wallet.LockAddr(a)
}

// Addr is the 20-byte form, the key of the memo topics; zero for a P2SH32
// address, which has no memo activity of its own (see attach.MemoAddresses).
func (a Address) Addr() [25]byte {
	return wallet.LockAddr(a).Addr()
}

func (a Address) IsP2SH32() bool {
	return wallet.LockAddr(a).IsP2SH32()
}

func (a Address) IsZero() bool {
	return wallet.LockAddr(a).IsZero()
}

// NewAddress wraps a 20-byte address.
func NewAddress(addr [25]byte) Address {
	return Address(wallet.Addr(addr).LockAddr())
}

func MarshalAddress(address Address) graphql.Marshaler {
	data, _ := json.Marshal(address.String())
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalAddress(v interface{}) (Address, error) {
	switch v := v.(type) {
	case string:
		addr, err := wallet.GetLockAddrFromString(v)
		if err != nil {
			return Address{}, fmt.Errorf("error unmarshal parsing string as address; %w", err)
		}
		return Address(addr), nil
	default:
		return Address{}, fmt.Errorf("error unmarshal unexpected address type not string: %T", v)
	}
}

// AddressesToLockAddrs unwraps a list argument.
func AddressesToLockAddrs(addresses []Address) []wallet.LockAddr {
	var lockAddrs = make([]wallet.LockAddr, len(addresses))
	for i := range addresses {
		lockAddrs[i] = addresses[i].LockAddr()
	}
	return lockAddrs
}
