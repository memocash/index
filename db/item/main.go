package item

import (
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
	"sort"
)

func GetTopics() []db.Object {
	return db.CombineObjects([]db.Object{
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
		&TxSuspect{},
	},
		addr.GetTopics(),
		chain.GetTopics(),
		memo.GetTopics(),
	)
}

func GetTopicsSorted() []db.Object {
	topics := GetTopics()
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].GetTopic() < topics[j].GetTopic()
	})
	return topics
}
