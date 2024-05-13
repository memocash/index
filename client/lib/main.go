package lib

import (
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Balance struct {
	Balance        int64
	Spendable      int64
	UtxoCount      int
	SpendableCount int
}

type Database interface {
	GetAddressBalance([]wallet.Addr) (*Balance, error)
	GetAddressLastUpdate([]wallet.Addr) ([]graph.AddressUpdate, error)
	GetUtxos([]wallet.Addr) ([]graph.Output, error)
	SaveTxs([]graph.Tx) error
	SetAddressLastUpdate([]graph.AddressUpdate) error
}
