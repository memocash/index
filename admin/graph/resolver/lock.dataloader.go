package resolver

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func LockLoader(ctx context.Context, addressString string) (*model.Lock, error) {
	address, err := wallet.GetAddrFromString(addressString)
	if err != nil {
		return nil, jerr.Getf(err, "error getting address from dataloader: %s", addressString)
	}
	var lock = &model.Lock{Address: address.String()}
	if HasField(ctx, "balance") {
		// TODO: Reimplement if needed
		return nil, jerr.New("error balance no longer implemented")
	}
	return lock, nil
}
