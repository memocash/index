package build

import (
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

type Wallet struct {
	Getter       gen.InputGetter
	FaucetGetter gen.InputGetter
	FaucetSaver  gen.FaucetSaver
	KeyRing      wallet.KeyRing
	Address      wallet.Address
	SlpAddress   wallet.Address
	OldAddress   wallet.Address
}

func (w Wallet) GetPkHash() []byte {
	return w.Address.GetPkHash()
}

func (w Wallet) GetSlpAddress() wallet.Address {
	if w.SlpAddress.IsSet() {
		return w.SlpAddress
	}
	return w.Address
}

func (w Wallet) GetAddresses() []wallet.Address {
	var addresses = []wallet.Address{w.Address}
	if w.SlpAddress.IsSet() {
		addresses = append(addresses, w.SlpAddress)
	}
	if w.OldAddress.IsSet() {
		addresses = append(addresses, w.OldAddress)
	}
	return addresses
}

func (w Wallet) GetChange() wallet.Change {
	return wallet.Change{
		Main: w.Address,
		Slp:  w.SlpAddress,
	}
}
