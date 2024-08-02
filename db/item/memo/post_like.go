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

type PostLike struct {
	PostTxHash [32]byte
	Seen       time.Time
	LikeTxHash [32]byte
	Addr       [25]byte
}

func (l *PostLike) GetTopic() string {
	return db.TopicMemoPostLike
}

func (l *PostLike) GetShardSource() uint {
	return client.GenShardSource(l.PostTxHash[:])
}

func (l *PostLike) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(l.PostTxHash[:]),
		jutil.GetTimeByteNanoBig(l.Seen),
		jutil.ByteReverse(l.LikeTxHash[:]),
	)
}

func (l *PostLike) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo liked")
	}
	copy(l.PostTxHash[:], jutil.ByteReverse(uid[:32]))
	l.Seen = jutil.GetByteTimeNanoBig(uid[32:40])
	copy(l.LikeTxHash[:], jutil.ByteReverse(uid[40:72]))
}

func (l *PostLike) Serialize() []byte {
	return l.Addr[:]
}

func (l *PostLike) Deserialize(data []byte) {
	if len(data) != memo.AddressLength {
		panic("invalid data size for memo liked")
	}
	copy(l.Addr[:], data)
}

func GetPostLikes(ctx context.Context, postTxHashes [][32]byte) ([]*PostLike, error) {
	messages, err := db.GetByPrefixes(ctx, db.TopicMemoPostLike, db.ShardPrefixesTxHashes(postTxHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message memo post likes; %w", err)
	}
	var postLikes = make([]*PostLike, len(messages))
	for i := range messages {
		postLikes[i] = new(PostLike)
		db.Set(postLikes[i], messages[i])
	}
	return postLikes, nil
}

func ListenPostLikes(ctx context.Context, postTxHashes [][32]byte) (chan *PostLike, error) {
	if len(postTxHashes) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range postTxHashes {
		shard := client.GenShardSource32(postTxHashes[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHashes[i][:]))
	}
	shardConfigs := config.GetQueueShards()
	var likedChan = make(chan *PostLike)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(likedChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoPostLike, prefixes)
		if err != nil {
			return nil, fmt.Errorf("error listening to db memo post liked by prefix; %w", err)
		}
		go func() {
			for msg := range chanMessage {
				var liked = new(PostLike)
				db.Set(liked, *msg)
				likedChan <- liked
			}
			cancelCtx.Cancel()
		}()
	}
	return likedChan, nil
}
