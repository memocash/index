package maint

import (
	"github.com/jchavannes/jgo/jerr"
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

func (c *CheckFollows) Check() error {
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		for {
			if err := dbClient.GetWOpts(client.Opts{
				Topic: db.TopicLockMemoFollow,
				Start: startUid,
				Max:   client.ExLargeLimit,
			}); err != nil {
				return jerr.Get("error getting db memo follow by prefix", err)
			}
			for _, msg := range dbClient.Messages {
				c.Processed++
				var lockMemoFollow = new(memo.LockFollow)
				db.Set(lockMemoFollow, msg)
				startUid = lockMemoFollow.GetUid()
				if len(lockMemoFollow.Follow) == 0 {
					c.BadFollows++
					if !c.Delete {
						continue
					}
					if err := memo.RemoveLockFollow(lockMemoFollow); err != nil {
						return jerr.Get("error removing lock memo follow", err)
					}
					var lockMemoFollowed = &memo.LockFollowed{
						FollowLockHash: lockMemoFollow.Follow,
						Height:         lockMemoFollow.Height,
						TxHash:         lockMemoFollow.TxHash,
						LockHash:       lockMemoFollow.LockHash,
						Unfollow:       lockMemoFollow.Unfollow,
					}
					if err := memo.RemoveLockFollowed(lockMemoFollowed); err != nil {
						return jerr.Get("error removing lock memo followed", err)
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
