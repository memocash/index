package wlt

import (
	"fmt"
	"github.com/memocash/index/client/lib"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Wallet struct {
	Key     wallet.PrivateKey
	Address wallet.Addr
	Getter  *InputGetter
}

func NewWallet(key wallet.PrivateKey, client *lib.Client) *Wallet {
	addr := key.GetAddr()
	return &Wallet{
		Key:     key,
		Address: addr,
		Getter:  NewInputGetter(addr, client),
	}
}

func (w *Wallet) BasicTx(script memo.Script) (*memo.Tx, error) {
	memoTx, err := gen.Tx(gen.TxRequest{
		Getter: w.Getter,
		Outputs: []*memo.Output{{
			Script: script,
		}},
		Change: wallet.Change{Main: w.Address.OldAddress()},
		KeyRing: wallet.KeyRing{
			Keys: []wallet.PrivateKey{w.Key},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error generating basic wallet tx; %w", err)
	}
	return memoTx, err
}
