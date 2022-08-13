package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/config"
)

type LockOutput struct {
	LockHash []byte
	Hash     []byte
	Index    uint32
}

func (o LockOutput) GetUid() []byte {
	return GetLockOutputUid(o.LockHash, o.Hash, o.Index)
}

func (o LockOutput) GetShard() uint {
	return client.GetByteShard(o.LockHash)
}

func (o LockOutput) GetTopic() string {
	return db.TopicLockOutput
}

func (o LockOutput) Serialize() []byte {
	return nil
}

func (o *LockOutput) SetUid(uid []byte) {
	if len(uid) != 68 {
		return
	}
	o.LockHash = uid[:32]
	o.Hash = jutil.ByteReverse(uid[32:64])
	o.Index = jutil.GetUint32(uid[64:68])
}

func (o *LockOutput) Deserialize([]byte) {}

func GetLockOutputUid(lockHash, hash []byte, index uint32) []byte {
	return jutil.CombineBytes(lockHash, jutil.ByteReverse(hash), jutil.GetUint32Data(index))
}

func GetLockOutputs(lockHash, start []byte) ([]*LockOutput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicLockOutput,
		Start:    start,
		Prefixes: [][]byte{lockHash},
		Max:      client.ExLargeLimit,
	})
	if err != nil {
		return nil, jerr.Get("error getting db lock outputs by prefix", err)
	}
	var lockOutputs = make([]*LockOutput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		lockOutputs[i] = new(LockOutput)
		db.Set(lockOutputs[i], dbClient.Messages[i])
	}
	return lockOutputs, nil
}

func GetLockOutputsSpecific(outs []memo.Out) ([]*LockOutput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := db.GetShardByte32(script.GetLockHash(out.PkScript))
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	var lockOutputs []*LockOutput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(outGroup))
		for i := range outGroup {
			prefixes[i] = jutil.CombineBytes(
				script.GetLockHash(outGroup[i].PkScript),
				jutil.ByteReverse(outGroup[i].TxHash),
				jutil.GetUint32Data(outGroup[i].Index),
			)
		}
		err := dbClient.GetByPrefixes(db.TopicLockOutput, prefixes)
		if err != nil {
			return nil, jerr.Get("error getting lock outputs by prefixes", err)
		}
		for i := range dbClient.Messages {
			var outputInput = new(LockOutput)
			db.Set(outputInput, dbClient.Messages[i])
			lockOutputs = append(lockOutputs, outputInput)
		}
	}
	return lockOutputs, nil
}
