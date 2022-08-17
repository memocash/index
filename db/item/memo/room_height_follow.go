package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type RoomHeightFollow struct {
	RoomHash []byte
	Height   int64
	TxHash   []byte
	Unfollow bool
	LockHash []byte
}

func (f RoomHeightFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.RoomHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(f.Height)),
		jutil.ByteReverse(f.TxHash),
	)
}

func (f RoomHeightFollow) GetShard() uint {
	return client.GetByteShard(f.RoomHash)
}

func (f RoomHeightFollow) GetTopic() string {
	return db.TopicMemoRoomHeightFollow
}

func (f RoomHeightFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.LockHash,
	)
}

func (f *RoomHeightFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	f.RoomHash = uid[:32]
	f.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	f.TxHash = jutil.ByteReverse(uid[40:72])
}

func (f *RoomHeightFollow) Deserialize(data []byte) {
	if len(data) < 1+memo.TxHashLength+1 {
		return
	}
	f.Unfollow = data[0] == 1
	f.LockHash = data[1:33]
}

func GetRoomHeightFollows(ctx context.Context, room string) ([]*RoomHeightFollow, error) {
	roomHash := GetRoomHash(room)
	shardConfig := config.GetShardConfig(client.GetByteShard32(roomHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoRoomHeightFollow,
		Prefixes: [][]byte{roomHash},
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo room height follows", err)
	}
	var roomHeightFollows = make([]*RoomHeightFollow, len(dbClient.Messages))
	for i := range dbClient.Messages {
		roomHeightFollows[i] = new(RoomHeightFollow)
		db.Set(roomHeightFollows[i], dbClient.Messages[i])
	}
	return roomHeightFollows, nil
}
