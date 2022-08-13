package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MemoPostRoom struct {
	TxHash []byte
	Room   string
}

func (r MemoPostRoom) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash)
}

func (r MemoPostRoom) GetShard() uint {
	return client.GetByteShard(r.TxHash)
}

func (r MemoPostRoom) GetTopic() string {
	return db.TopicMemoPostRoom
}

func (r MemoPostRoom) Serialize() []byte {
	return []byte(r.Room)
}

func (r *MemoPostRoom) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	r.TxHash = jutil.ByteReverse(uid)
}

func (r *MemoPostRoom) Deserialize(data []byte) {
	r.Room = string(data)
}
