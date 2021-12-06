package balance

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/node/obj/get"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Balance struct {
	Address   string
	Balance   int64
	Spendable int64
	Spends    int
	UtxoCount int
	Outputs   int
	Txs       int
}

func (b *Balance) Get(addressString string) error {
	address := wallet.GetAddressFromString(addressString)
	var lockScript []byte
	var err error
	if !address.IsSet() {
		return jerr.New("error parsing address")
	} else if address.IsP2PKH() {
		s := script.P2pkh{PkHash: address.GetPkHash()}
		lockScript, err = s.Get()
		if err != nil {
			return jerr.Get("error getting lock script for p2pkh address", err)
		}
	} else if address.IsP2SH() {
		s := script.P2sh{ScriptHash: address.ScriptAddress()}
		lockScript, err = s.Get()
		if err != nil {
			return jerr.Get("error getting lock script for p2sh address", err)
		}
	} else {
		return jerr.Newf("error unknown address type")
	}
	b.Address = address.GetEncoded()
	balance := get.NewBalance(lockScript)
	if err = balance.GetBalance(); err != nil {
		return jerr.Get("error getting balance", err)
	}
	b.Balance = balance.Balance
	b.Spendable = balance.Spendable
	b.UtxoCount = balance.UtxoCount
	b.Spends = balance.Spends
	return nil
}

func NewBalance() *Balance {
	return &Balance{}
}
