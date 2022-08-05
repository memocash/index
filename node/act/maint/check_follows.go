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
				Topic: item.TopicLockMemoFollow,
				Start: startUid,
				Max:   client.ExLargeLimit,
			}); err != nil {
				return jerr.Get("error getting db memo follow by prefix", err)
			}
			for _, msg := range db.Messages {
				c.Processed++
				var lockMemoFollow = new(item.LockMemoFollow)
				item.Set(lockMemoFollow, msg)
				startUid = lockMemoFollow.GetUid()
				if len(lockMemoFollow.Follow) == 0 {
					c.BadFollows++
					if !c.Delete {
						continue
					}
					if err := item.RemoveLockMemoFollow(lockMemoFollow); err != nil {
						return jerr.Get("error removing lock memo follow", err)
					}
					var lockMemoFollowed = &item.LockMemoFollowed{
						FollowLockHash: lockMemoFollow.Follow,
						Height:         lockMemoFollow.Height,
						TxHash:         lockMemoFollow.TxHash,
						LockHash:       lockMemoFollow.LockHash,
						Unfollow:       lockMemoFollow.Unfollow,
					}
					if err := item.RemoveLockMemoFollowed(lockMemoFollowed); err != nil {
						return jerr.Get("error removing lock memo followed", err)
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
