package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/graph/model"
)

func GetLock(ctx context.Context, address model.Address) (*model.Lock, error) {
	var lock = &model.Lock{Address: address}
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
