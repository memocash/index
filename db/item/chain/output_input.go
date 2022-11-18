package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type OutputInput struct {
	PrevHash  [32]byte
	PrevIndex uint32
	Hash      [32]byte
	Index     uint32
}

func (t *OutputInput) GetTopic() string {
	return db.TopicOutputInput
}

func (t *OutputInput) GetShard() uint {
	return client.GetByteShard(t.PrevHash[:])
}

func (t *OutputInput) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(t.PrevHash[:]),
		jutil.GetUint32DataBig(t.PrevIndex),
		jutil.ByteReverse(t.Hash[:]),
		jutil.GetUint32DataBig(t.Index),
	)
}

func (t *OutputInput) SetUid(uid []byte) {
	if len(uid) != 72 {
		return
	}
	copy(t.PrevHash[:], jutil.ByteReverse(uid[:32]))
	t.PrevIndex = jutil.GetUint32Big(uid[32:36])
	copy(t.Hash[:], jutil.ByteReverse(uid[36:68]))
	t.Index = jutil.GetUint32Big(uid[68:72])
}

func (t *OutputInput) Serialize() []byte {
	return nil
}

func (t *OutputInput) Deserialize([]byte) {}
