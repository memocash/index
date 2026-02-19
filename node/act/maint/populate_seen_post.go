package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/config"
)

type PopulateSeenPost struct {
	Posts int
	Ctx   context.Context
}

func NewPopulateSeenPost(ctx context.Context) *PopulateSeenPost {
	return &PopulateSeenPost{
		Ctx: ctx,
	}
}

func (p *PopulateSeenPost) Populate() error {
	for _, shardConfig := range config.GetQueueShards() {
		var startTxHash [32]byte
		dbClient := client.NewClient(shardConfig.GetHost())
		for {
			opt := client.OptionHugeLimit()
			if err := dbClient.GetByPrefix(p.Ctx, db.TopicMemoPost, client.NewStart(startTxHash[:]), opt); err != nil {
				return fmt.Errorf("error getting memo posts for populate seen posts; %w", err)
			}
			var postTxHashes [][32]byte
			for _, dbPost := range dbClient.Messages {
				var post = new(memo.Post)
				db.Set(post, dbPost)
				postTxHashes = append(postTxHashes, post.TxHash)
				if jutil.ByteGT(post.TxHash[:], startTxHash[:]) {
					startTxHash = post.TxHash
				}
			}
			seenTxs, err := chain.GetTxSeens(p.Ctx, postTxHashes)
			if err != nil {
				return fmt.Errorf("error getting tx seens for populate seen posts; %w", err)
			}
			var objects []db.Object
			for _, seenTx := range seenTxs {
				objects = append(objects, &memo.SeenPost{
					Seen:       seenTx.Timestamp,
					PostTxHash: seenTx.TxHash,
				})
			}
			if err := db.Save(objects); err != nil {
				return fmt.Errorf("error saving seen posts; %w", err)
			}
			p.Posts += len(objects)
			if len(dbClient.Messages) < client.HugeLimit {
				break
			}
		}
	}
	return nil
}
