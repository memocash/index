package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

func GetLock(ctx context.Context, addressString string) (*model.Lock, error) {
	address, err := wallet.GetAddrFromString(addressString)
	if err != nil {
		return nil, fmt.Errorf("error getting address from dataloader: %s; %w", addressString, err)
	}
	var lock = &model.Lock{Address: address.String()}
	fields := GetFields(ctx)
	if err := AttachToLocks(ctx, fields, []*model.Lock{lock}); err != nil {
		return nil, fmt.Errorf("error attaching details to lock; %w", err)
	}
	if fields.HasField("balance") {
		// TODO: Reimplement if needed
		return nil, fmt.Errorf("error balance no longer implemented")
	}
	return lock, nil
}
