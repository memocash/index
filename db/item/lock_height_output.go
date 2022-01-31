package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

type LockHeightOutput struct {
	LockHash []byte
	Height   int64
	Hash     []byte
	Index    uint32
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

func GetLockHeightOutputUid(lockHash []byte, height int64, hash []byte, index uint32) []byte {
	return jutil.CombineBytes(
		lockHash,
		jutil.GetInt64DataBig(height),
		jutil.ByteReverse(hash),
		jutil.GetUint32Data(index),
	)
}
