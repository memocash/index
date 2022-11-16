package item

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
)

const HeightMempool = -1

type LockHeightOutput struct {
	LockHash []byte
	Height   int64
	Hash     []byte
	Index    uint32
}

func (o LockHeightOutput) IsMempool() bool {
	return o.Height == HeightMempool
}

func (o LockHeightOutput) GetUid() []byte {
	return GetLockHeightOutputUid(o.LockHash, o.Height, o.Hash, o.Index)
}

func (o LockHeightOutput) GetShard() uint {
	return client.GetByteShard(o.LockHash)
}

func (o LockHeightOutput) GetTopic() string {
	return db.TopicLockHeightOutput
}

func (o LockHeightOutput) Serialize() []byte {
	return nil
}

func (o *LockHeightOutput) SetUid(uid []byte) {
	if len(uid) != 76 {
		return
	}
	o.LockHash = uid[:32]
	o.Height = jutil.GetInt64Big(uid[32:40])
	o.Hash = jutil.ByteReverse(uid[40:72])
	o.Index = jutil.GetUint32(uid[72:76])
}

func (o *LockHeightOutput) Deserialize([]byte) {}

func ListenMempoolLockHeightOutputs(ctx context.Context, lockHash []byte) (chan *LockHeightOutput, error) {
	lockHeightChan, err := ListenMempoolLockHeightOutputsMultiple(ctx, [][]byte{lockHash})
	if err != nil {
		return nil, jerr.Get("error getting lock height output listen message chan", err)
	}
	if len(lockHeightChan) != 1 {
		return nil, jerr.Newf("invalid lock height output listen message chan length: %d", len(lockHeightChan))
	}
	return lockHeightChan[0], nil
}

func ListenMempoolLockHeightOutputsMultiple(ctx context.Context, lockHashes [][]byte) ([]chan *LockHeightOutput, error) {
	var shardLockHashGroups = make(map[uint32][][]byte)
	for _, lockHash := range lockHashes {
		shard := db.GetShardByte32(lockHash)
		shardLockHashGroups[shard] = append(shardLockHashGroups[shard], lockHash)
	}
	var chanLockHeightOutputs []chan *LockHeightOutput
	for shard, lockHashGroup := range shardLockHashGroups {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(ctx, db.TopicLockHeightOutput, lockHashGroup)
		if err != nil {
			return nil, jerr.Get("error getting lock height output listen message chan", err)
		}
		var chanLockHeightOutput = make(chan *LockHeightOutput)
		go func() {
			for {
				msg, ok := <-chanMessage
				if !ok {
					close(chanLockHeightOutput)
					return
				}
				var lockHeightOutput = new(LockHeightOutput)
				db.Set(lockHeightOutput, *msg)
				chanLockHeightOutput <- lockHeightOutput
			}
		}()
		chanLockHeightOutputs = append(chanLockHeightOutputs, chanLockHeightOutput)
	}
	return chanLockHeightOutputs, nil
}

func GetLockHeightOutputs(lockHash, start []byte) ([]*LockHeightOutput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicLockHeightOutput,
		Start:    start,
		Prefixes: [][]byte{lockHash},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db lock outputs by prefix", err)
	}
	var lockHeightOutputs = make([]*LockHeightOutput, len(dbClient.Messages))
	for i := range dbClient.Messages {
		lockHeightOutputs[i] = new(LockHeightOutput)
		db.Set(lockHeightOutputs[i], dbClient.Messages[i])
	}
	return lockHeightOutputs, nil
}

func GetLockHeightOutputUid(lockHash []byte, height int64, hash []byte, index uint32) []byte {
	return jutil.CombineBytes(
		lockHash,
		jutil.GetInt64DataBig(height),
		jutil.ByteReverse(hash),
		jutil.GetUint32Data(index),
	)
}

func RemoveLockHeightOutputs(lockHeightOutputs []*LockHeightOutput) error {
	var shardUidsMap = make(map[uint32][][]byte)
	for _, lockHeightOutput := range lockHeightOutputs {
		shard := db.GetShard32(lockHeightOutput.GetShard())
		shardUidsMap[shard] = append(shardUidsMap[shard], lockHeightOutput.GetUid())
	}
	for shard, shardUids := range shardUidsMap {
		shardUids = jutil.RemoveDupesAndEmpties(shardUids)
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.DeleteMessages(db.TopicLockHeightOutput, shardUids); err != nil {
			return jerr.Get("error deleting items topic lock height output", err)
		}
	}
	return nil
}
