package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/script"
	"github.com/memocash/server/ref/config"
)

type LockOutput struct {
	LockHash []byte
	Hash     []byte
	Index    uint32
}

func (o LockOutput) GetUid() []byte {
	return jutil.CombineBytes(o.LockHash, jutil.ByteReverse(o.Hash), jutil.GetUint32Data(o.Index))
}

func (o LockOutput) GetShard() uint {
	return client.GetByteShard(o.LockHash)
}

func (o LockOutput) GetTopic() string {
	return TopicLockOutput
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

func GetLockOutputs(lockHash, start []byte) ([]*LockOutput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	err := db.GetWOpts(client.Opts{
		Topic:    TopicLockOutput,
		Start:    start,
		Prefixes: [][]byte{lockHash},
		Max:      client.ExLargeLimit,
	})
	if err != nil {
		return nil, jerr.Get("error getting db lock outputs by prefix", err)
	}
	var lockOutputs = make([]*LockOutput, len(db.Messages))
	for i := range db.Messages {
		lockOutputs[i] = new(LockOutput)
		lockOutputs[i].SetUid(db.Messages[i].Uid)
		lockOutputs[i].Deserialize(db.Messages[i].Message)
	}
	return lockOutputs, nil
}

func GetLockOutputsSpecific(outs []memo.Out) ([]*LockOutput, error) {
	var shardOutGroups = make(map[uint32][]memo.Out)
	for _, out := range outs {
		shard := GetShardByte32(script.GetLockHash(out.PkScript))
		shardOutGroups[shard] = append(shardOutGroups[shard], out)
	}
	var lockOutputs []*LockOutput
	for shard, outGroup := range shardOutGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		var prefixes = make([][]byte, len(outGroup))
		for i := range outGroup {
			prefixes[i] = jutil.CombineBytes(
				script.GetLockHash(outGroup[i].PkScript),
				jutil.ByteReverse(outGroup[i].TxHash),
				jutil.GetUint32Data(outGroup[i].Index),
			)
		}
		err := db.GetByPrefixes(TopicLockOutput, prefixes)
		if err != nil {
			return nil, jerr.Get("error getting lock outputs by prefixes", err)
		}
		for i := range db.Messages {
			var outputInput = new(LockOutput)
			outputInput.SetUid(db.Messages[i].Uid)
			outputInput.Deserialize(db.Messages[i].Message)
			lockOutputs = append(lockOutputs, outputInput)
		}
	}
	return lockOutputs, nil
}
