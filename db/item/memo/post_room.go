package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type PostRoom struct {
	TxHash []byte
	Room   string
}

func (r PostRoom) GetUid() []byte {
	return jutil.ByteReverse(r.TxHash)
}

func (r PostRoom) GetShard() uint {
	return client.GetByteShard(r.TxHash)
}

func (r PostRoom) GetTopic() string {
	return db.TopicMemoPostRoom
}

func (r PostRoom) Serialize() []byte {
	return []byte(r.Room)
}

func (r *PostRoom) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	r.TxHash = jutil.ByteReverse(uid)
}

func (r *PostRoom) Deserialize(data []byte) {
	r.Room = string(data)
}
