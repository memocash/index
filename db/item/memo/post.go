package memo

import (
	"context"
	"fmt"
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

func (p *Post) GetShardSource() uint {
	return client.GenShardSource(p.TxHash[:])
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

func GetPost(ctx context.Context, txHash [32]byte) (*Post, error) {
	posts, err := GetPosts(ctx, [][32]byte{txHash})
	if err != nil {
		return nil, fmt.Errorf("error getting memo posts for single; %w", err)
	}
	if len(posts) == 0 {
		return nil, nil
	}
	return posts[0], nil
}

func GetPosts(ctx context.Context, txHashes [][32]byte) ([]*Post, error) {
	messages, err := db.GetSpecific(ctx, db.TopicMemoPost, db.ShardUidsTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting client message memo posts; %w", err)
	}
	var posts = make([]*Post, len(messages))
	for i := range messages {
		posts[i] = new(Post)
		db.Set(posts[i], messages[i])
	}
	return posts, nil
}

func ListenPosts(ctx context.Context) (chan *Post, error) {
	var postChan = make(chan *Post)
	cancelCtx := db.NewCancelContext(ctx, func() {
		close(postChan)
	})
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		chanMessage, err := dbClient.Listen(cancelCtx.Context, db.TopicMemoPost, nil)
		if err != nil {
			return nil, fmt.Errorf("error listening to db memo posts (all); %w", err)
		}
		go func() {
			for msg := range chanMessage {
				var post = new(Post)
				db.Set(post, *msg)
				postChan <- post
			}
			cancelCtx.Cancel()
		}()
	}
	return postChan, nil
}
