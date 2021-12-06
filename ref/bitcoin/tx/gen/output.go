package gen

import (
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func GetAddressOutput(address wallet.Address, quantity int64) *memo.Output {
	if address.IsP2SH() {
		return &memo.Output{
			Script: &script.P2sh{ScriptHash: address.ScriptAddress()},
			Amount: quantity,
		}
	} else {
		return &memo.Output{
			Script: &script.P2pkh{PkHash: address.GetPkHash()},
			Amount: quantity,
		}
	}
}
