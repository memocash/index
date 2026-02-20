package item

import (
	"sort"

	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/db/item/slp"
)

func GetTopics() []db.Object {
	return db.CombineObjects([]db.Object{
		&FoundPeer{},
		&Message{},
		&Peer{},
		&PeerConnection{},
		&PeerFound{},
		&ProcessError{},
		&ProcessStatus{},
		&SyncStatus{},
	},
		addr.GetTopics(),
		chain.GetTopics(),
		memo.GetTopics(),
		slp.GetTopics(),
	)
}

func GetTopicsSorted() []db.Object {
	topics := GetTopics()
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].GetTopic() < topics[j].GetTopic()
	})
	return topics
}
