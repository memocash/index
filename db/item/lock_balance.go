package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

type LockBalance struct {
	LockHash  []byte
	Balance   int64
	Spendable int64
	UtxoCount int
	Spends    int
}

func (b LockBalance) GetUid() []byte {
	return b.LockHash
}

func (b LockBalance) GetShard() uint {
	return client.GetByteShard(b.LockHash)
}

func (b LockBalance) GetTopic() string {
	return TopicLockBalance
}

func (b LockBalance) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.GetInt64Data(b.Balance),
		jutil.GetInt64Data(b.Spendable),
		jutil.GetIntData(b.UtxoCount),
		jutil.GetIntData(b.Spends),
	)
}

func (b *LockBalance) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	b.LockHash = uid[:32]
}

func (b *LockBalance) Deserialize(data []byte) {
	if len(data) < 24 {
		return
	}
	b.Balance = jutil.GetInt64(data[:8])
	b.Spendable = jutil.GetInt64(data[8:16])
	b.UtxoCount = jutil.GetInt(data[16:20])
	b.Spends = jutil.GetInt(data[20:24])
}

func GetLockBalance(lockHash []byte) (*LockBalance, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	err := db.GetSingle(TopicLockBalance, lockHash)
	if err != nil && !client.IsMessageNotSetError(err) {
		return nil, jerr.Get("error getting db lock balance single", err)
	}
	if len(db.Messages) != 1 {
		return nil, jerr.Get("error lock balance not found", client.EntryNotFoundError)
	}
	var lockBalance = new(LockBalance)
	lockBalance.SetUid(db.Messages[0].Uid)
	lockBalance.Deserialize(db.Messages[0].Message)
	return lockBalance, nil
}

func RemoveLockBalances(lockHashes [][]byte) error {
	lockHashes = jutil.RemoveDupesAndEmpties(lockHashes)
	var shardUidsMap = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := GetShardByte32(lockHash)
		shardUidsMap[shard] = append(shardUidsMap[shard], lockHash)
	}
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.DeleteMessages(TopicLockBalance, shardUids); err != nil {
			return jerr.Get("error deleting topic lock balances", err)
		}
	}
	return nil
}
