package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type LockRoomFollow struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	Room     string
}

func (f LockRoomFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.LockHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash),
	)
}

func (f LockRoomFollow) GetShard() uint {
	return client.GetByteShard(f.LockHash)
}

func (f LockRoomFollow) GetTopic() string {
	return db.TopicLockMemoRoomFollow
}

func (f LockRoomFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		[]byte(f.Room),
	)
}

func (f *LockRoomFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength*2+memo.Int8Size {
		return
	}
	f.LockHash = uid[:32]
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	f.TxHash = jutil.ByteReverse(uid[40:72])
}

func (f *LockRoomFollow) Deserialize(data []byte) {
	if len(data) < 1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.Room = string(data[1:])
}
