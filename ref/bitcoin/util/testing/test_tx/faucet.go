package test_tx

import (
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

func GetFaucetSaverWithKey(key string) gen.FaucetSaver {
	return FaucetSaver{
		Key: key,
	}
}

type FaucetSaver struct {
	Key string
}

func (f FaucetSaver) Save([]byte, []byte, []byte, []byte) error {
	return nil
}

func (f FaucetSaver) IsFreeTx(outputs []*memo.Output) bool {
	return memo.IsFreeTx(outputs)
}

func (f FaucetSaver) GetKey() wallet.PrivateKey {
	faucetKey, _ := wallet.ImportPrivateKey(f.Key)
	return faucetKey
}
