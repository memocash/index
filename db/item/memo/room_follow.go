package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type RoomFollow struct {
	RoomHash []byte
	Seen     time.Time
	TxHash   [32]byte
	Unfollow bool
	Addr     [25]byte
}

func (f *RoomFollow) GetTopic() string {
	return db.TopicMemoRoomFollow
}

func (f *RoomFollow) GetShardSource() uint {
	return client.GenShardSource(f.RoomHash)
}

func (f *RoomFollow) GetUid() []byte {
	return jutil.CombineBytes(
		f.RoomHash,
		jutil.GetTimeByteNanoBig(f.Seen),
		jutil.ByteReverse(f.TxHash[:]),
	)
}

func (f *RoomFollow) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	f.RoomHash = uid[:32]
	f.Seen = jutil.GetByteTimeNanoBig(uid[32:40])
	copy(f.TxHash[:], jutil.ByteReverse(uid[40:72]))
}

func (f *RoomFollow) Serialize() []byte {
	var unfollow byte
	if f.Unfollow {
		unfollow = 1
	}
	return jutil.CombineBytes(
		[]byte{unfollow},
		f.Addr[:],
	)
}

func (f *RoomFollow) Deserialize(data []byte) {
	if len(data) < 1+memo.AddressLength+1 {
		return
	}
	f.Unfollow = data[0] == 1
	copy(f.Addr[:], data[1:26])
}

func GetRoomFollows(ctx context.Context, room string) ([]*RoomFollow, error) {
	roomHash := GetRoomHash(room)
	shardConfig := config.GetShardConfig(client.GenShardSource32(roomHash), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoRoomFollow,
		Prefixes: [][]byte{roomHash},
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, fmt.Errorf("error getting db memo room follows; %w", err)
	}
	var roomFollows = make([]*RoomFollow, len(dbClient.Messages))
	for i := range dbClient.Messages {
		roomFollows[i] = new(RoomFollow)
		db.Set(roomFollows[i], dbClient.Messages[i])
	}
	return roomFollows, nil
}
