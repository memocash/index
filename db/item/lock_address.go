package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

type LockAddress struct {
	LockHash []byte
	Address  string
}

func (a LockAddress) GetUid() []byte {
	return a.LockHash
}

func (a LockAddress) GetShard() uint {
	return client.GetByteShard(a.LockHash)
}

func (a LockAddress) GetTopic() string {
	return db.TopicLockAddress
}

func (a LockAddress) Serialize() []byte {
	return []byte(a.Address)
}

func (a *LockAddress) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	a.LockHash = uid[:32]
}

func (a *LockAddress) Deserialize(data []byte) {
	a.Address = string(data)
}

func GetLockAddress(lockHash []byte) (*LockAddress, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetSingle(db.TopicLockAddress, lockHash); err != nil && !client.IsMessageNotSetError(err) {
		return nil, jerr.Get("error getting db lock address single", err)
	}
	if len(dbClient.Messages) != 1 {
		return nil, jerr.Get("error lock address not found", client.EntryNotFoundError)
	}
	var lockAddress = new(LockAddress)
	db.Set(lockAddress, dbClient.Messages[0])
	return lockAddress, nil
}

func GetLockAddresses(lockHashes [][]byte) ([]*LockAddress, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := uint32(db.GetShardByte(lockHash))
		shardPrefixes[shard] = append(shardPrefixes[shard], lockHash)
	}
	var lockAddresses []*LockAddress
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicLockAddress, prefixes); err != nil {
			return nil, jerr.Get("error getting db message tx outputs", err)
		}
		for _, msg := range dbClient.Messages {
			var lockAddress = new(LockAddress)
			db.Set(lockAddress, msg)
			lockAddresses = append(lockAddresses, lockAddress)
		}
	}
	return lockAddresses, nil
}
