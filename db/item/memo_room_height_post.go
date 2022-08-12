package item

import (
	"context"
	"crypto/sha256"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
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

func GetMemoRoomHeightPosts(ctx context.Context, room string) ([]*MemoRoomHeightPost, error) {
	roomHash := GetMemoRoomHash(room)
	shard := client.GetByteShard32(roomHash)
	shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
	db := client.NewClient(shardConfig.GetHost())
	if err := db.GetWOpts(client.Opts{
		Topic:    TopicMemoRoomHeightPost,
		Prefixes: [][]byte{roomHash},
		Max:      client.ExLargeLimit,
		Context:  ctx,
	}); err != nil {
		return nil, jerr.Get("error getting db memo room height posts", err)
	}
	var memoRoomHeightPosts []*MemoRoomHeightPost
	for _, msg := range db.Messages {
		var memoRoomHeightPost = new(MemoRoomHeightPost)
		Set(memoRoomHeightPost, msg)
		memoRoomHeightPosts = append(memoRoomHeightPosts, memoRoomHeightPost)
	}
	return memoRoomHeightPosts, nil
}
