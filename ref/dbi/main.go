package dbi

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type TxSave interface {
	SaveTxs(*Block) error
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

func GetBlock(block *wire.MsgBlock) *Block {
	return &Block{
		Header:       block.Header,
		Transactions: block.Transactions,
	}
}

func GetBlockWithHeight(block *wire.MsgBlock, height int64) *Block {
	return &Block{
		Header:       block.Header,
		Transactions: block.Transactions,
		Height:       height,
	}
}

func GetBlockSingleTx(tx *wire.MsgTx) *Block {
	return GetBlock(memo.GetBlockFromTxs([]*wire.MsgTx{tx}, nil))
}

type Block struct {
	Header       wire.BlockHeader
	Transactions []*wire.MsgTx
	Height       int64
}

func (b Block) Msg() *wire.MsgBlock {
	return &wire.MsgBlock{
		Header:       b.Header,
		Transactions: b.Transactions,
	}
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
