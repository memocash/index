package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockUtxoLost struct {
	LockHash []byte
	Hash     []byte
	Index    uint32
	Value    int64
	Special  bool
}

func (o LockUtxoLost) GetUid() []byte {
	return GetLockOutputUid(o.LockHash, o.Hash, o.Index)
}

func (o LockUtxoLost) GetShard() uint {
	return client.GetByteShard(o.LockHash)
}

func (o LockUtxoLost) GetTopic() string {
	return db.TopicLockUtxoLost
}

func (o LockUtxoLost) Serialize() []byte {
	var special byte
	if o.Special {
		special = 1
	}
	return jutil.CombineBytes(
		jutil.GetInt64Data(o.Value),
		[]byte{special},
	)
}

func (o *LockUtxoLost) SetUid(uid []byte) {
	if len(uid) != 68 {
		return
	}
	o.LockHash = uid[:32]
	o.Hash = jutil.ByteReverse(uid[32:64])
	o.Index = jutil.GetUint32(uid[64:68])
}

func (o *LockUtxoLost) Deserialize(data []byte) {
	if len(data) < 9 {
		return
	}
	o.Value = jutil.GetInt64(data[:8])
	o.Special = data[8] == 1
}

func GetLockUtxoLosts(outs []memo.Out) ([]*LockUtxoLost, error) {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := uint32(db.GetShard(db.GetShardByte(out.LockHash)))
		shardUidsMap[shard] = append(shardUidsMap[shard], GetLockOutputUid(out.LockHash, out.TxHash, out.Index))
	}
	var lockUtxoLosts []*LockUtxoLost
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic: db.TopicLockUtxoLost,
			Uids:  shardUids,
		}); err != nil {
			return nil, jerr.Get("error getting db lock utxo losts by outs", err)
		}
		for i := range dbClient.Messages {
			var lockUtxoLost = new(LockUtxoLost)
			db.Set(lockUtxoLost, dbClient.Messages[i])
			lockUtxoLosts = append(lockUtxoLosts, lockUtxoLost)
		}
	}
	return lockUtxoLosts, nil
}

func RemoveLockUtxoLosts(lockUtxos []*LockUtxoLost) error {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, lockUtxo := range lockUtxos {
		shard := uint32(db.GetShard(lockUtxo.GetShard()))
		shardUidsMap[shard] = append(shardUidsMap[shard], lockUtxo.GetUid())
	}
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.DeleteMessages(db.TopicLockUtxoLost, shardUids); err != nil {
			return jerr.Get("error deleting topic lock utxos", err)
		}
	}
	return nil
}
