package load

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/admin/graph/model"
)

func GetTx(ctx context.Context, txHash [32]byte) (*model.Tx, error) {
	var tx = &model.Tx{Hash: txHash}
	if err := AttachToTxs(ctx, GetFields(ctx), []*model.Tx{tx}); err != nil {
		return nil, fmt.Errorf("error attaching all to single tx; %w", err)
	}
	return tx, nil
}

func GetTxByString(ctx context.Context, txHash string) (*model.Tx, error) {
	txHashBytes, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, fmt.Errorf("error decoding tx hash from string for graph load; %w", err)
	}
	tx, err := GetTx(ctx, *txHashBytes)
	if err != nil {
		return nil, fmt.Errorf("error getting tx for graph load from string; %w", err)
	}
	return tx, nil
}
