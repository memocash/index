package memo

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type RoomFollow struct {
	RoomHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	LockHash []byte
}

func (f RoomFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.RoomHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash),
	)
}

func (f RoomFollow) GetShard() uint {
	return client.GetByteShard(f.RoomHash)
}

func (f RoomFollow) GetTopic() string {
	return db.TopicMemoRoomFollow
}

func (f RoomFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.LockHash,
	)
}

func (f *RoomFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	f.RoomHash = uid[:32]
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	f.TxHash = jutil.ByteReverse(uid[40:72])
}

func (f *RoomFollow) Deserialize(data []byte) {
	if len(data) < 1+memo.TxHashLength+1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.LockHash = data[1:33]
}
