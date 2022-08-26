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

type PostHeightLike struct {
	PostTxHash []byte
	Height     int64
	LikeTxHash []byte
	LockHash   []byte
}

func (l PostHeightLike) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(l.PostTxHash),
		jutil.ByteFlip(jutil.GetInt64DataBig(l.Height)),
		jutil.ByteReverse(l.LikeTxHash),
	)
}

func (l PostHeightLike) GetShard() uint {
	return client.GetByteShard(l.PostTxHash)
}

func (l PostHeightLike) GetTopic() string {
	return db.TopicMemoPostHeightLike
}

func (l PostHeightLike) Serialize() []byte {
	return l.LockHash
}

func (l *PostHeightLike) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo liked")
	}
	l.PostTxHash = jutil.ByteReverse(uid[:32])
	l.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	l.LikeTxHash = jutil.ByteReverse(uid[40:72])
}

func (l *PostHeightLike) Deserialize(data []byte) {
	if len(data) != memo.LockHashLength {
		panic("invalid data size for memo liked")
	}
	l.LockHash = data
}

func GetPostHeightLikes(postTxHashes [][]byte) ([]*PostHeightLike, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := db.GetShardByte32(postTxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash))
	}
	var likeds []*PostHeightLike
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicMemoPostHeightLike, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo likeds", err)
		}
		for _, msg := range dbClient.Messages {
			var liked = new(PostHeightLike)
			db.Set(liked, msg)
			likeds = append(likeds, liked)
		}
	}
	return likeds, nil
}

func ListenPostHeightLikes(ctx context.Context, postTxHashes [][]byte) (chan *PostHeightLike, error) {
	if len(postTxHashes) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := client.GetByteShard32(postTxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash))
	}
	shardConfigs := config.GetQueueShards()
	var likedChan = make(chan *PostHeightLike)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(likedChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoPostHeightLike, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo post liked by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var liked = new(PostHeightLike)
				db.Set(liked, *msg)
				likedChan <- liked
			}
			cancelCtx.Cancel()
		}()
	}
	return likedChan, nil
}
