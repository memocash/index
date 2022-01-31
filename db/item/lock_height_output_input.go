package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
)

type LockHeightOutputInput struct {
	LockHash  []byte
	Height    int64
	PrevHash  []byte
	PrevIndex uint32
	Hash      []byte
	Index     uint32
}

func (t LockHeightOutputInput) GetUid() []byte {
	return jutil.CombineBytes(
		t.LockHash,
		jutil.GetInt64DataBig(t.Height),
		jutil.ByteReverse(t.PrevHash),
		jutil.GetUint32Data(t.PrevIndex),
		jutil.ByteReverse(t.Hash),
		jutil.GetUint32Data(t.Index),
	)
}

func (t *LockHeightOutputInput) SetUid(uid []byte) {
	if len(uid) != 112 {
		return
	}
	t.LockHash = uid[:32]
	t.Height = jutil.GetInt64Big(uid[32:40])
	t.PrevHash = jutil.ByteReverse(uid[40:72])
	t.PrevIndex = jutil.GetUint32(uid[72:76])
	t.Hash = jutil.ByteReverse(uid[76:108])
	t.Index = jutil.GetUint32(uid[108:112])
}

func (t LockHeightOutputInput) GetShard() uint {
	return client.GetByteShard(t.PrevHash)
}

func (t LockHeightOutputInput) GetTopic() string {
	return TopicLockHeightOutputInput
}

func (t LockHeightOutputInput) Serialize() []byte {
	return nil
}

func (t *LockHeightOutputInput) Deserialize([]byte) {}
