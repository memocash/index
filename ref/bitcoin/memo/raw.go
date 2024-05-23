package memo

import (
	"bytes"
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"math"
	"time"
)

func GetBlockFromTxs(txs []*wire.MsgTx, header *wire.BlockHeader) *wire.MsgBlock {
	var block = &wire.MsgBlock{
		Transactions: txs,
	}
	if header != nil {
		block.Header = *header
	} else {
		block.Header.Timestamp = time.Unix(0, 0)
	}
	return block
}

func GetRaw(msg *wire.MsgTx) []byte {
	writer := new(bytes.Buffer)
	msg.Serialize(writer)
	return writer.Bytes()
}

func GetMsgFromRaw(raw []byte) (*wire.MsgTx, error) {
	msgTx := wire.NewMsgTx(1)
	reader := bytes.NewReader(raw)
	err := msgTx.Deserialize(reader)
	if err != nil {
		return nil, fmt.Errorf("error deserializing tx; %w", err)
	}
	return msgTx, nil
}

func GetRawBlockHeader(blockHeader wire.BlockHeader) []byte {
	writer := new(bytes.Buffer)
	blockHeader.Serialize(writer)
	return writer.Bytes()
}

func GetBlockHeaderFromRaw(raw []byte) (*wire.BlockHeader, error) {
	var blockHeader = new(wire.BlockHeader)
	reader := bytes.NewReader(raw)
	err := blockHeader.Deserialize(reader)
	if err != nil {
		return nil, fmt.Errorf("error deserializing block header; %w", err)
	}
	return blockHeader, nil
}

func GetRawBlock(block wire.MsgBlock) []byte {
	writer := new(bytes.Buffer)
	block.Serialize(writer)
	return writer.Bytes()
}

func GetBlockFromRaw(raw []byte) (*wire.MsgBlock, error) {
	var block = new(wire.MsgBlock)
	reader := bytes.NewReader(raw)
	err := block.Deserialize(reader)
	if err != nil {
		return nil, fmt.Errorf("error deserializing block; %w", err)
	}
	return block, nil
}

func IsCoinbase(prevHash []byte, prevIndex uint32) bool {
	return jutil.AllZeros(prevHash) && prevIndex == math.MaxUint32
}

func IsCoinbaseInput(in *wire.TxIn) bool {
	return IsCoinbase(in.PreviousOutPoint.Hash.CloneBytes(), in.PreviousOutPoint.Index)
}

func HasCoinbase(tx *wire.MsgTx) bool {
	for _, in := range tx.TxIn {
		if IsCoinbaseInput(in) {
			return true
		}
	}
	return false
}
