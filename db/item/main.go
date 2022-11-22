package item

import (
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
)

func GetTopics() []db.Object {
	return append([]db.Object{
		&DoubleSpendInput{},
		&DoubleSpendOutput{},
		&DoubleSpendSeen{},
		&FoundPeer{},
		&HeightBlockShard{},
		&HeightProcessed{},
		&LockAddress{},
		&LockBalance{},
		&LockHeightOutput{},
		&LockHeightOutputInput{},
		&LockOutput{},
		&LockUtxo{},
		&LockUtxoLost{},
		&MempoolTxRaw{},
		&Message{},
		&Peer{},
		&PeerConnection{},
		&PeerFound{},
		&ProcessError{},
		&ProcessStatus{},
		&TxLost{},
		&TxProcessed{},
		&TxSeen{},
		&TxSuspect{},
	},
		append(chain.GetTopics(), memo.GetTopics()...)...,
	)
}
