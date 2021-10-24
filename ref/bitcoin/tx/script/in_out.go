package script

import (
	"encoding/binary"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/wallet"
)

type InOut interface {
	Get() [][]byte
	Size() int64
}

type InOutInput struct {
	TxHash []byte
	Index  uint32
}

func (i InOutInput) Get() [][]byte {
	var indexBytes = make([]byte, 2)
	binary.BigEndian.PutUint16(indexBytes, uint16(i.Index))
	return [][]byte{
		{memo.InOutTypeInput},
		i.TxHash,
		indexBytes,
	}
}

func (i InOutInput) Size() int64 {
	return getBsSize(i.Get())
}

type InOutOutput struct {
	Address  wallet.Address
	IsToken  bool
	IsSelf   bool
	Quantity uint64
}

func (i InOutOutput) Get() [][]byte {
	var returnBytes [][]byte
	if i.IsSelf {
		if i.IsToken {
			returnBytes = append(returnBytes, []byte{memo.InOutTypeTokenOutputSelf})
		} else {
			returnBytes = append(returnBytes, []byte{memo.InOutTypeBitcoinOutputSelf})
		}
	} else if ! i.Address.IsP2SH() {
		if i.IsToken {
			returnBytes = append(returnBytes, []byte{memo.InOutTypeTokenOutputP2pkh})
		} else {
			returnBytes = append(returnBytes, []byte{memo.InOutTypeBitcoinOutputP2pkh})
		}
	} else {
		if i.IsToken {
			returnBytes = append(returnBytes, []byte{memo.InOutTypeTokenOutputP2sh})
		} else {
			returnBytes = append(returnBytes, []byte{memo.InOutTypeBitcoinOutputP2sh})
		}
	}
	if ! i.IsSelf {
		returnBytes = append(returnBytes, i.Address.GetPkHash())
	}
	var quantityBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(quantityBytes, i.Quantity)
	returnBytes = append(returnBytes, quantityBytes)
	return returnBytes
}

func (i InOutOutput) Size() int64 {
	return getBsSize(i.Get())
}

func getBsSize(bs [][]byte) int64 {
	var totalSize int64
	for _, b := range bs {
		totalSize += len64(b)
	}
	return totalSize + int64(len(bs))*memo.OutputOpDataFee
}
