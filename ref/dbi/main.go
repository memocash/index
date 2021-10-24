package dbi

import (
	"github.com/jchavannes/btcd/wire"
)

type TxSave interface {
	SaveTxs(*wire.MsgBlock) error
}

type BlockSave interface {
	SaveBlock(wire.BlockHeader) error
	GetBlock(int64) ([]byte, error)
}

type BlockHeightSave interface {
	SaveHeights([]*BlockHeight) error
}

type TxBlockSave interface {
	SaveTxBlocks([]*TxBlock) error
}

type InputSave interface {
	SaveInputs([]*Input) error
}

type OutputSave interface {
	SaveOutputs([]*Output) error
}

type Input struct {
	Hash      []byte
	Index     uint32
	PrevHash  []byte
	PrevIndex uint32
	LockHash  []byte
}

type Output struct {
	LockHash []byte
	Hash     []byte
	Index    uint32
	Value    int64
}

type TxBlock struct {
	TxHash    []byte
	BlockHash []byte
}

type BlockHeight struct {
	BlockHash []byte
	Height    int64
}
