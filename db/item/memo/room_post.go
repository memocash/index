package memo

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type RoomPost struct {
	RoomHash []byte
	Seen     time.Time
	TxHash   [32]byte
}

func (p *RoomPost) GetTopic() string {
	return db.TopicMemoRoomPost
}

func (p *RoomPost) GetShardSource() uint {
	return client.GenShardSource(p.RoomHash)
}

func (p *RoomPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.RoomHash,
		jutil.GetTimeByteNanoBig(p.Seen),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *RoomPost) SetUid(uid []byte) {
	if len(uid) < memo.TxHashLength*2+memo.Int8Size {
		return
	}
	p.RoomHash = uid[:32]
	p.Seen = jutil.GetByteTimeNanoBig(uid[32:40])
	copy(p.TxHash[:], jutil.ByteReverse(uid[40:72]))
}

func (p *RoomPost) Serialize() []byte {
	return nil
}

func (p *RoomPost) Deserialize([]byte) {}

func GetRoomHash(room string) []byte {
	sum := sha256.Sum256([]byte(room))
	return sum[:]
}

func GetRoomPosts(ctx context.Context, room string) ([]*RoomPost, error) {
	roomHash := GetRoomHash(room)
	dbClient := db.GetShardClient(client.GenShardSource32(roomHash))
	if err := dbClient.GetByPrefix(ctx, db.TopicMemoRoomPost, client.NewPrefix(roomHash)); err != nil {
		return nil, fmt.Errorf("error getting db memo room posts; %w", err)
	}
	var roomPosts = make([]*RoomPost, len(dbClient.Messages))
	for i := range dbClient.Messages {
		roomPosts[i] = new(RoomPost)
		db.Set(roomPosts[i], dbClient.Messages[i])
	}
	return roomPosts, nil
}

func ListenRoomPosts(ctx context.Context, rooms []string) (chan *RoomPost, error) {
	if len(rooms) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, room := range rooms {
		roomHash := GetRoomHash(room)
		shard := client.GenShardSource32(roomHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], roomHash)
	}
	shardConfigs := config.GetQueueShards()
	var roomPostChan = make(chan *RoomPost)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(roomPostChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoRoomPost, prefixes)
		if err != nil {
			return nil, fmt.Errorf("error listening to db memo room post by prefix; %w", err)
		}
		go func() {
			for msg := range chanMessage {
				var roomPost = new(RoomPost)
				db.Set(roomPost, *msg)
				roomPostChan <- roomPost
			}
			cancelCtx.Cancel()
		}()
	}
	return roomPostChan, nil
}
