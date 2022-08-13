package item

import (
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
)

func GetTopics() []db.Object {
	return append([]db.Object{
		&Block{},
		&BlockHeight{},
		&BlockTx{},
		&DoubleSpendInput{},
		&DoubleSpendOutput{},
		&DoubleSpendSeen{},
		&FoundPeer{},
		&HeightBlock{},
		&HeightBlockShard{},
		&HeightDuplicate{},
		&HeightProcessed{},
		&LockAddress{},
		&LockBalance{},
		&LockHeightOutput{},
		&LockHeightOutputInput{},
		&LockOutput{},
		&LockUtxo{},
		&LockUtxoLost{},
		&LockMemoFollow{},
		&LockMemoFollowed{},
		&LockMemoLike{},
		&LockMemoName{},
		&LockMemoPost{},
		&LockMemoProfile{},
		&LockMemoProfilePic{},
		&MempoolTxRaw{},
		&Message{},
		&OutputInput{},
		&Peer{},
		&PeerConnection{},
		&PeerFound{},
		&ProcessError{},
		&ProcessStatus{},
		&Tx{},
		&TxBlock{},
		&TxInput{},
		&TxLost{},
		&TxOutput{},
		&TxProcessed{},
		&TxSeen{},
		&TxSuspect{},
	},
		memo.GetTopics()...,
	)
}
