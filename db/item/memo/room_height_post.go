package memo

import (
	"context"
	"crypto/sha256"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type RoomHeightPost struct {
	RoomHash []byte
	Height   int64
	TxHash   []byte
}

func (p RoomHeightPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.RoomHash,
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash),
	)
}

func (p RoomHeightPost) GetShard() uint {
	return client.GetByteShard(p.RoomHash)
}

func (p RoomHeightPost) GetTopic() string {
	return db.TopicMemoRoomHeightPost
}

func (p RoomHeightPost) Serialize() []byte {
	return nil
}

func (p *RoomHeightPost) SetUid(uid []byte) {
	if len(uid) < memo.TxHashLength*2+memo.Int8Size {
		return
	}
	p.RoomHash = uid[:32]
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	p.TxHash = jutil.ByteReverse(uid[40:72])
}

func (p *RoomHeightPost) Deserialize([]byte) {}

func GetRoomHash(room string) []byte {
	sum := sha256.Sum256([]byte(room))
	return sum[:]
}

func GetRoomHeightPosts(ctx context.Context, room string) ([]*RoomHeightPost, error) {
	roomHash := GetRoomHash(room)
	shard := client.GetByteShard32(roomHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:    db.TopicMemoRoomHeightPost,
		Prefixes: [][]byte{roomHash},
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo room height posts", err)
	}
	var roomHeightPosts = make([]*RoomHeightPost, len(dbClient.Messages))
	for i := range dbClient.Messages {
		roomHeightPosts[i] = new(RoomHeightPost)
		db.Set(roomHeightPosts[i], dbClient.Messages[i])
	}
	return roomHeightPosts, nil
}

func ListenRoomPosts(ctx context.Context, rooms []string) (chan *RoomHeightPost, error) {
	if len(rooms) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, room := range rooms {
		roomHash := GetRoomHash(room)
		shard := client.GetByteShard32(roomHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], roomHash)
	}
	shardConfigs := config.GetQueueShards()
	var roomHeightPostChan = make(chan *RoomHeightPost)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(roomHeightPostChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoRoomHeightPost, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo room height post by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var roomHeightPost = new(RoomHeightPost)
				db.Set(roomHeightPost, *msg)
				roomHeightPostChan <- roomHeightPost
			}
			cancelCtx.Cancel()
		}()
	}
	return roomHeightPostChan, nil
}
