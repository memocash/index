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

type Liked struct {
	PostTxHash []byte
	Height     int64
	LikeTxHash []byte
	LockHash   []byte
}

func (l Liked) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(l.PostTxHash),
		jutil.ByteFlip(jutil.GetInt64DataBig(l.Height)),
		jutil.ByteReverse(l.LikeTxHash),
	)
}

func (l Liked) GetShard() uint {
	return client.GetByteShard(l.PostTxHash)
}

func (l Liked) GetTopic() string {
	return db.TopicMemoLiked
}

func (l Liked) Serialize() []byte {
	return l.LockHash
}

func (l *Liked) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		panic("invalid uid size for memo liked")
	}
	l.PostTxHash = jutil.ByteReverse(uid[:32])
	l.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[32:40]))
	l.LikeTxHash = jutil.ByteReverse(uid[40:72])
}

func (l *Liked) Deserialize(data []byte) {
	if len(data) != memo.LockHashLength {
		panic("invalid data size for memo liked")
	}
	l.LockHash = data
}

func GetLikeds(postTxHashes [][]byte) ([]*Liked, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := db.GetShardByte32(postTxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash))
	}
	var likeds []*Liked
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicMemoLiked, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo likeds", err)
		}
		for _, msg := range dbClient.Messages {
			var liked = new(Liked)
			db.Set(liked, msg)
			likeds = append(likeds, liked)
		}
	}
	return likeds, nil
}

func ListenLikeds(ctx context.Context, postTxHashes [][]byte) (chan *Liked, error) {
	if len(postTxHashes) == 0 {
		return nil, nil
	}
	var shardPrefixes = make(map[uint32][][]byte)
	for _, postTxHash := range postTxHashes {
		shard := client.GetByteShard32(postTxHash)
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(postTxHash))
	}
	shardConfigs := config.GetQueueShards()
	var likedChan = make(chan *Liked)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(likedChan)
	})
	for shard, prefixes := range shardPrefixes {
		dbClient := client.NewClient(config.GetShardConfig(shard, shardConfigs).GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoLiked, prefixes)
		if err != nil {
			return nil, jerr.Get("error listening to db memo post liked by prefix", err)
		}
		go func() {
			for msg := range chanMessage {
				var liked = new(Liked)
				db.Set(liked, *msg)
				likedChan <- liked
			}
			cancelCtx.Cancel()
		}()
	}
	return likedChan, nil
}
