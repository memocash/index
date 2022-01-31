package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

type LockMempoolOutput struct {
	LockHash []byte
	Hash     []byte
	Index    uint32
}

func (o LockMempoolOutput) GetUid() []byte {
	return GetLockOutputUid(o.LockHash, o.Hash, o.Index)
}

func (o LockMempoolOutput) GetShard() uint {
	return client.GetByteShard(o.LockHash)
}

func (o LockMempoolOutput) GetTopic() string {
	return TopicLockMempoolOutput
}

func (o LockMempoolOutput) Serialize() []byte {
	return nil
}

func (o *LockMempoolOutput) SetUid(uid []byte) {
	if len(uid) != 68 {
		return
	}
	o.LockHash = uid[:32]
	o.Hash = jutil.ByteReverse(uid[32:64])
	o.Index = jutil.GetUint32(uid[64:68])
}

func (o *LockMempoolOutput) Deserialize([]byte) {}
