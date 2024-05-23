package model

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"io"
)

type Address [25]byte

func (a Address) String() string {
	return wallet.Addr(a).String()
}

func MarshalAddress(address Address) graphql.Marshaler {
	data, _ := json.Marshal(wallet.Addr(address).String())
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(data))
	})
}

func UnmarshalAddress(v interface{}) (Address, error) {
	switch v := v.(type) {
	case string:
		addr, err := wallet.GetAddrFromString(v)
		if err != nil {
			return Address{}, fmt.Errorf("error unmarshal parsing string as address; %w", err)
		}
		return Address(*addr), nil
	default:
		return Address{}, fmt.Errorf("error unmarshal unexpected address type not string: %T", v)
	}
}

func AddressesToArrays(addresses []Address) [][25]byte {
	var addressArrays [][25]byte
	for _, address := range addresses {
		addressArrays = append(addressArrays, address)
	}
	return addressArrays
}
