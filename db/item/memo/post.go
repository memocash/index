package memo

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type Post struct {
	TxHash [32]byte
	Addr   [25]byte
	Post   string
}

func (p *Post) GetTopic() string {
	return db.TopicMemoPost
}

func (p *Post) GetShard() uint {
	return client.GetByteShard(p.TxHash[:])
}

func (p *Post) GetUid() []byte {
	return jutil.ByteReverse(p.TxHash[:])
}

func (p *Post) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(p.TxHash[:], jutil.ByteReverse(uid))
}

func (p *Post) Serialize() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		[]byte(p.Post),
	)
}

func (p *Post) Deserialize(data []byte) {
	if len(data) < memo.AddressLength {
		return
	}
	copy(p.Addr[:], data[:25])
	p.Post = string(data[25:])
}

func GetPost(txHash [32]byte) (*Post, error) {
	posts, err := GetPosts([][32]byte{txHash})
	if err != nil {
		return nil, jerr.Get("error getting memo posts for single", err)
	}
	if len(posts) == 0 {
		return nil, nil
	}
	return posts[0], nil
}

func GetPosts(txHashes [][32]byte) ([]*Post, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardByte32(txHash[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], jutil.ByteReverse(txHash[:]))
	}
	var posts []*Post
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, config.GetQueueShards())
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetByPrefixes(db.TopicMemoPost, prefixes); err != nil {
			return nil, jerr.Get("error getting client message memo posts", err)
		}
		for _, msg := range dbClient.Messages {
			var post = new(Post)
			db.Set(post, msg)
			posts = append(posts, post)
		}
	}
	return posts, nil
}
