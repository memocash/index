package lib

import (
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"time"
)

type Balance struct {
	Balance        int64
	Spendable      int64
	UtxoCount      int
	SpendableCount int
}

type Database interface {
	GetAddressBalance(wallet.Addr) (*Balance, error)
	GetAddressLastUpdate(wallet.Addr) (time.Time, error)
	GetUtxos(wallet.Addr) ([]graph.Output, error)
	SaveTxs([]graph.Tx) error
	SetAddressLastUpdate(wallet.Addr, time.Time) error
}
