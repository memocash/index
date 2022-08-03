package maint

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
)

type CheckFollows struct {
	Delete     bool
	Processed  int
	BadFollows int
}

func (c *CheckFollows) Check() error {
	for _, shardConfig := range config.GetQueueShards() {
		db := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		for {
			if err := db.GetWOpts(client.Opts{
				Topic: item.TopicMemoFollow,
				Start: startUid,
				Max:   client.ExLargeLimit,
			}); err != nil {
				return jerr.Get("error getting db memo follow by prefix", err)
			}
			for _, msg := range db.Messages {
				c.Processed++
				var memoFollow = new(item.MemoFollow)
				item.Set(memoFollow, msg)
				startUid = memoFollow.GetUid()
				if len(memoFollow.Follow) == 0 {
					c.BadFollows++
					if !c.Delete {
						continue
					}
					if err := item.RemoveMemoFollow(memoFollow); err != nil {
						return jerr.Get("error removing memo follow", err)
					}
					var memoFollowed = &item.MemoFollowed{
						FollowLockHash: memoFollow.Follow,
						Height:         memoFollow.Height,
						TxHash:         memoFollow.TxHash,
						LockHash:       memoFollow.LockHash,
						Unfollow:       memoFollow.Unfollow,
					}
					if err := item.RemoveMemoFollowed(memoFollowed); err != nil {
						return jerr.Get("error removing memo followed", err)
					}
				}
			}
			if len(db.Messages) < client.ExLargeLimit {
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
