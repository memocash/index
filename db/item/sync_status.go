package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

const (
	SyncStatusSaveTxs        = "txs"
	SyncStatusProcessInitial = "process-initial"
)

type SyncStatus struct {
	Name   string
	Height int64
}

func (s *SyncStatus) GetUid() []byte {
	return []byte(s.Name)
}

func (s *SyncStatus) GetShard() uint {
	return client.GetByteShard([]byte(s.Name))
}

func (s *SyncStatus) GetTopic() string {
	return db.TopicSyncStatus
}

func (s *SyncStatus) Serialize() []byte {
	return jutil.GetInt64DataBig(s.Height)
}

func (s *SyncStatus) SetUid(uid []byte) {
	s.Name = string(uid)
}

func (s *SyncStatus) Deserialize(data []byte) {
	if len(data) != 8 {
		return
	}
	s.Height = jutil.GetInt64Big(data)
}

func GetSyncStatus(name string) (*SyncStatus, error) {
	var syncStatus = &SyncStatus{Name: name}
	if err := db.GetItem(syncStatus); err != nil {
		return nil, jerr.Get("error getting item sync status", err)
	}
	return syncStatus, nil
}
