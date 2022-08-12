package item

import (
	"crypto/sha256"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MemoRoomHeightPost struct {
	RoomHash []byte
	Height   int64
	TxHash   []byte
}

func (r MemoRoomHeightPost) GetUid() []byte {
	return jutil.CombineBytes(
		r.RoomHash,
		jutil.GetInt64Data(r.Height),
		jutil.ByteReverse(r.TxHash),
	)
}

func (r MemoRoomHeightPost) GetShard() uint {
	return client.GetByteShard(r.RoomHash)
}

func (r MemoRoomHeightPost) GetTopic() string {
	return TopicMemoRoomHeightPost
}

func (r MemoRoomHeightPost) Serialize() []byte {
	return nil
}

func (r *MemoRoomHeightPost) SetUid(uid []byte) {
	if len(uid) < memo.TxHashLength*2+memo.Int8Size {
		return
	}
	r.RoomHash = uid[:32]
	r.Height = jutil.GetInt64(uid[32:40])
	r.TxHash = jutil.ByteReverse(uid[40:72])
}

func (r *MemoRoomHeightPost) Deserialize([]byte) {}

func GetMemoRoomHash(room string) []byte {
	sum := sha256.Sum256([]byte(room))
	return sum[:]
}
