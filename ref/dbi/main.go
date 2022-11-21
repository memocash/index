package dbi

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
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

type Block struct {
	Header       wire.BlockHeader
	Transactions []Tx
}

func (b *Block) IsNil() bool {
	return b == nil || (b.Header.Timestamp.IsZero() && len(b.Transactions) == 0)
}

func (b *Block) ToWireBlock() *wire.MsgBlock {
	return BlockToWireBlock(b)
}

type Tx struct {
	BlockIndex uint64
	Version    int32
	LockTime   uint32
	Inputs     []Input
	Outputs    []Output
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
	block := &Block{Header: wireBlock.Header}
	for i, wireTx := range wireBlock.Transactions {
		tx := WireTxToTx(wireTx, uint64(i))
		block.Transactions = append(block.Transactions, *tx)
	}
	return block
}

func WireTxToTx(wireTx *wire.MsgTx, index uint64) *Tx {
	tx := &Tx{
		Version:    wireTx.Version,
		LockTime:   wireTx.LockTime,
		BlockIndex: index,
	}
	for _, wireInput := range wireTx.TxIn {
		input := Input{
			PrevHash:  wireInput.PreviousOutPoint.Hash.CloneBytes(),
			PrevIndex: wireInput.PreviousOutPoint.Index,
			Sequence:  wireInput.Sequence,
		}
		tx.Inputs = append(tx.Inputs, input)
	}
	for _, wireOutput := range wireTx.TxOut {
		output := Output{
			Value:      wireOutput.Value,
			LockScript: wireOutput.PkScript,
		}
		tx.Outputs = append(tx.Outputs, output)
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
	wireTx := wire.NewMsgTx(tx.Version)
	wireTx.LockTime = tx.LockTime
	for _, input := range tx.Inputs {
		wireTx.AddTxIn(&wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  *input.GetPrevHash(),
				Index: input.PrevIndex,
			},
			SignatureScript: input.UnlockScript,
			Sequence:        input.Sequence,
		})
	}
	for _, output := range tx.Outputs {
		wireTx.AddTxOut(&wire.TxOut{
			Value:    output.Value,
			PkScript: output.LockScript,
		})
	}
	return wireTx
}
