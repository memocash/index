package dbi

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/ref/bitcoin/memo"
	"time"
)

type TxSave interface {
	SaveTxs(*Block) error
}

type BlockSave interface {
	SaveBlock(BlockInfo) error
	GetBlock(int64) (*chainhash.Hash, error)
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

type Block struct {
	Header       wire.BlockHeader
	Height       int64
	Seen         time.Time
	Transactions []Tx
}

func (b *Block) IsNil() bool {
	return b == nil || (!b.HasHeader() && len(b.Transactions) == 0)
}

func (b *Block) HasHeader() bool {
	return BlockHeaderSet(b.Header)
}

func (b *Block) Size() int64 {
	n := int64(memo.BlockHeaderLength + wire.VarIntSerializeSize(uint64(len(b.Transactions))))
	for _, tx := range b.Transactions {
		n += int64(tx.MsgTx.SerializeSize())
	}
	return n
}

func BlockHeaderSet(header wire.BlockHeader) bool {
	return !jutil.IsTimeZero(header.Timestamp)
}

type BlockInfo struct {
	Header  wire.BlockHeader
	Size    int64
	TxCount int
}

type Tx struct {
	BlockIndex uint32
	Hash       [32]byte
	Seen       time.Time
	Saved      bool
	MsgTx      *wire.MsgTx
}

type Input struct {
	Hash         []byte
	Index        uint32
	Sequence     uint32
	PrevHash     []byte
	PrevIndex    uint32
	UnlockScript []byte
}

func (i *Input) GetPrevHash() *chainhash.Hash {
	h, _ := chainhash.NewHash(i.PrevHash)
	return h
}

type Output struct {
	Hash       []byte
	Index      uint32
	Value      int64
	LockScript []byte
}

type TxBlock struct {
	TxHash    []byte
	BlockHash []byte
}

type BlockHeight struct {
	BlockHash []byte
	Height    int64
}

func WireBlockToBlock(wireBlock *wire.MsgBlock) *Block {
	if wireBlock == nil {
		return &Block{}
	}
	block := &Block{
		Header: wireBlock.Header,
		Seen:   time.Now(),
	}
	if block.HasHeader() && block.Header.Timestamp.Before(block.Seen) {
		block.Seen = block.Header.Timestamp
	}
	for i, wireTx := range wireBlock.Transactions {
		tx := WireTxToTx(wireTx, uint32(i))
		tx.Seen = block.Seen
		block.Transactions = append(block.Transactions, *tx)
	}
	return block
}

func WireTxToTx(wireTx *wire.MsgTx, index uint32) *Tx {
	tx := &Tx{
		MsgTx:      wireTx,
		Hash:       wireTx.TxHash(),
		BlockIndex: index,
	}
	return tx
}

func BlockToWireBlock(block *Block) *wire.MsgBlock {
	if block == nil {
		return nil
	}
	wireBlock := wire.NewMsgBlock(&block.Header)
	for _, tx := range block.Transactions {
		wireBlock.AddTransaction(TxToWireTx(&tx))
	}
	return wireBlock
}

func TxToWireTx(tx *Tx) *wire.MsgTx {
	return tx.MsgTx
}
