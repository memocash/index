package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

type LockMempoolOutputInput struct {
	LockHash  []byte
	PrevHash  []byte
	PrevIndex uint32
	Hash      []byte
	Index     uint32
}

func (t LockMempoolOutputInput) GetUid() []byte {
	return jutil.CombineBytes(
		t.LockHash,
		jutil.ByteReverse(t.PrevHash),
		jutil.GetUint32Data(t.PrevIndex),
		jutil.ByteReverse(t.Hash),
		jutil.GetUint32Data(t.Index),
	)
}

func (t *LockMempoolOutputInput) SetUid(uid []byte) {
	if len(uid) != 104 {
		return
	}
	t.LockHash = uid[:32]
	t.PrevHash = jutil.ByteReverse(uid[32:64])
	t.PrevIndex = jutil.GetUint32(uid[64:68])
	t.Hash = jutil.ByteReverse(uid[68:100])
	t.Index = jutil.GetUint32(uid[100:104])
}

func (t LockMempoolOutputInput) GetShard() uint {
	return client.GetByteShard(t.PrevHash)
}

func (t LockMempoolOutputInput) GetTopic() string {
	return TopicLockMempoolOutputInput
}

func (t LockMempoolOutputInput) Serialize() []byte {
	return nil
}

func (t *LockMempoolOutputInput) Deserialize([]byte) {}
