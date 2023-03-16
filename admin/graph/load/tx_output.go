package load

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/model"
)

func GetTxOutput(ctx context.Context, txHash [32]byte, index uint32) (*model.TxOutput, error) {
	var txOutput = &model.TxOutput{Hash: txHash, Index: index}
	if err := AttachToOutputs(GetPreloads(ctx), []*model.TxOutput{txOutput}); err != nil {
		return nil, jerr.Get("error attaching all to single tx output", err)
	}
	return txOutput, nil
}

func GetTxOutputString(ctx context.Context, txHash string, index uint32) (*model.TxOutput, error) {
	txHashBytes, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, jerr.Get("error decoding tx hash from string for graph load tx output", err)
	}
	txOutput, err := GetTxOutput(ctx, *txHashBytes, index)
	if err != nil {
		return nil, jerr.Get("error getting tx output for graph load from string", err)
	}
	return txOutput, nil
}
