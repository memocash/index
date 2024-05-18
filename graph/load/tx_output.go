package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/graph/model"
)

func GetTxOutput(ctx context.Context, txHash [32]byte, index uint32) (*model.TxOutput, error) {
	var txOutput = &model.TxOutput{Hash: txHash, Index: index}
	if err := AttachToOutputs(ctx, GetFields(ctx), []*model.TxOutput{txOutput}); err != nil {
		return nil, fmt.Errorf("error attaching all to single tx output; %w", err)
	}
	return txOutput, nil
}
