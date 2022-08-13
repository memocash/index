package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type LockUtxo struct {
	LockHash []byte
	Hash     []byte
	Index    uint32
	Value    int64
	Special  bool
}

func (o LockUtxo) GetUid() []byte {
	return GetLockOutputUid(o.LockHash, o.Hash, o.Index)
}

func (o LockUtxo) GetShard() uint {
	return client.GetByteShard(o.LockHash)
}

func (o LockUtxo) GetTopic() string {
	return db.TopicLockUtxo
}

func (o LockUtxo) Serialize() []byte {
	var special byte
	if o.Special {
		special = 1
	}
	return jutil.CombineBytes(
		jutil.GetInt64Data(o.Value),
		[]byte{special},
	)
}

func (o *LockUtxo) SetUid(uid []byte) {
	if len(uid) != 68 {
		return
	}
	o.LockHash = uid[:32]
	o.Hash = jutil.ByteReverse(uid[32:64])
	o.Index = jutil.GetUint32(uid[64:68])
}

func (o *LockUtxo) Deserialize(data []byte) {
	if len(data) < 9 {
		return
	}
	o.Value = jutil.GetInt64(data[:8])
	o.Special = data[8] == 1
}

func GetLockUtxos(lockHash, start []byte) ([]*LockUtxo, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicLockUtxo,
		Start:    start,
		Prefixes: [][]byte{lockHash},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db lock utxos by prefix", err)
	}
	var lockOutputs = make([]*LockUtxo, len(dbClient.Messages))
	for i := range dbClient.Messages {
		lockOutputs[i] = new(LockUtxo)
		db.Set(lockOutputs[i], dbClient.Messages[i])
	}
	return lockOutputs, nil
}

func GetLockUtxosByOuts(outs []memo.Out) ([]*LockUtxo, error) {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, out := range outs {
		shard := uint32(db.GetShard(db.GetShardByte(out.LockHash)))
		shardUidsMap[shard] = append(shardUidsMap[shard], GetLockOutputUid(out.LockHash, out.TxHash, out.Index))
	}
	var lockUtxos []*LockUtxo
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic: db.TopicLockUtxo,
			Uids:  shardUids,
		}); err != nil {
			return nil, jerr.Get("error getting db lock utxos by outs", err)
		}
		for i := range dbClient.Messages {
			var lockUtxoLost = new(LockUtxo)
			db.Set(lockUtxoLost, dbClient.Messages[i])
			lockUtxos = append(lockUtxos, lockUtxoLost)
		}
	}
	return lockUtxos, nil
}

func RemoveLockUtxos(lockUtxos []*LockUtxo) error {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, lockUtxo := range lockUtxos {
		shard := uint32(db.GetShard(lockUtxo.GetShard()))
		shardUidsMap[shard] = append(shardUidsMap[shard], lockUtxo.GetUid())
	}
	for shard, shardUids := range shardUidsMap {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.DeleteMessages(db.TopicLockUtxo, shardUids); err != nil {
			return jerr.Get("error deleting topic lock utxos", err)
		}
	}
	return nil
}
