package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
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

func (l *PostLike) GetShard() uint {
	return client.GetByteShard(l.PostTxHash[:])
}

func (l *PostLike) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(l.PostTxHash[:]),
		jutil.GetTimeByteBig(l.Seen),
		jutil.ByteReverse(l.LikeTxHash[:]),
	)
}

func (l *PostLike) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo liked")
	}
	copy(l.PostTxHash[:], jutil.ByteReverse(uid[:32]))
	l.Seen = jutil.GetByteTimeBig(uid[32:40])
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

func GetPostLikes(postTxHashes [][32]byte) ([]*PostLike, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := db.GetShardByte32(postTxHash[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash[:]))
	}
	var likeds []*PostLike
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicMemoPostLike, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo likeds", err)
		}
		for _, msg := range dbClient.Messages {
			var liked = new(PostLike)
			db.Set(liked, msg)
			likeds = append(likeds, liked)
		}
	}
	return likeds, nil
}

func ListenPostLikes(ctx context.Context, postTxHashes [][32]byte) (chan *PostLike, error) {
	if len(postTxHashes) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := client.GetByteShard32(postTxHash[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash[:]))
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
			return nil, jerr.Get("error listening to db memo post liked by prefix", err)
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
