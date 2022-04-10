package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
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
	return TopicLockHeightOutput
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

func ListenMempoolLockHeightOutputs(lockHash []byte) (chan *LockHeightOutput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	prefix := jutil.CombineBytes(
		lockHash,
		jutil.GetInt64DataBig(HeightMempool),
	)
	chanMessage, err := db.Listen(TopicLockHeightOutput, [][]byte{prefix})
	if err != nil {
		return nil, jerr.Get("error getting lock height output listen message chan", err)
	}
	var chanLockHeightOutput = make(chan *LockHeightOutput)
	go func() {
		for {
			msg := <-chanMessage
			if msg == nil {
				chanLockHeightOutput <- nil
				close(chanLockHeightOutput)
				return
			}
			var lockHeightOutput = new(LockHeightOutput)
			lockHeightOutput.SetUid(msg.Uid)
			lockHeightOutput.Deserialize(msg.Message)
			chanLockHeightOutput <- lockHeightOutput
		}
	}()
	return chanLockHeightOutput, nil
}

func GetLockHeightOutputs(lockHash, start []byte) ([]*LockHeightOutput, error) {
	shardConfig := config.GetShardConfig(client.GetByteShard32(lockHash), config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicLockHeightOutput,
		Start:    start,
		Prefixes: [][]byte{lockHash},
		Max:      client.ExLargeLimit,
	}); err != nil {
		return nil, jerr.Get("error getting db lock outputs by prefix", err)
	}
	var lockHeightOutputs = make([]*LockHeightOutput, len(db.Messages))
	for i := range db.Messages {
		lockHeightOutputs[i] = new(LockHeightOutput)
		lockHeightOutputs[i].SetUid(db.Messages[i].Uid)
		lockHeightOutputs[i].Deserialize(db.Messages[i].Message)
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
		shard := GetShard32(lockHeightOutput.GetShard())
		shardUidsMap[shard] = append(shardUidsMap[shard], lockHeightOutput.GetUid())
	}
	for shard, shardUids := range shardUidsMap {
		shardUids = jutil.RemoveDupesAndEmpties(shardUids)
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		db := client.NewClient(shardConfig.GetHost())
		if err := db.DeleteMessages(TopicLockHeightOutput, shardUids); err != nil {
			return jerr.Get("error deleting items topic lock height output", err)
		}
	}
	return nil
}
