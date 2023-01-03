package lib

import (
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Database interface {
	GetAddressBalance(address *wallet.Addr) (int64, error)
	GetAddressHeight(address *wallet.Addr) (int64, error)
	SaveTxs(txs []graph.Tx) error
}
