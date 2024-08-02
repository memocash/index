package maint

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/config"
)

type CheckFollows struct {
	Delete     bool
	Processed  int
	BadFollows int
}

func (c *CheckFollows) Check(ctx context.Context) error {
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		for {
			opt := client.OptionExLargeLimit()
			if err := dbClient.GetByPrefix(ctx, db.TopicMemoAddrFollow, client.NewStart(startUid), opt); err != nil {
				return fmt.Errorf("error getting db memo follow by prefix; %w", err)
			}
			for _, msg := range dbClient.Messages {
				c.Processed++
				var addrMemoFollow = new(memo.AddrFollow)
				db.Set(addrMemoFollow, msg)
				startUid = addrMemoFollow.GetUid()
				if jutil.AllZeros(addrMemoFollow.FollowAddr[:]) {
					c.BadFollows++
					if !c.Delete {
						continue
					}
					var addrMemoFollowed = &memo.AddrFollowed{
						FollowAddr: addrMemoFollow.FollowAddr,
						Seen:       addrMemoFollow.Seen,
						TxHash:     addrMemoFollow.TxHash,
						Addr:       addrMemoFollow.Addr,
						Unfollow:   addrMemoFollow.Unfollow,
					}
					if err := db.Remove([]db.Object{addrMemoFollow, addrMemoFollowed}); err != nil {
						return fmt.Errorf("error removing addr memo follow/followed; %w", err)
					}
				}
			}
			if len(dbClient.Messages) < client.ExLargeLimit {
				break
			}
		}
	}
	return nil
}

func NewCheckFollows(delete bool) *CheckFollows {
	return &CheckFollows{
		Delete: delete,
	}
}
