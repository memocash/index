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
				Topic: db.TopicMemoAddrHeightFollow,
				Start: startUid,
				Max:   client.ExLargeLimit,
			}); err != nil {
				return jerr.Get("error getting db memo follow by prefix", err)
			}
			for _, msg := range dbClient.Messages {
				c.Processed++
				var addrMemoFollow = new(memo.AddrHeightFollow)
				db.Set(addrMemoFollow, msg)
				startUid = addrMemoFollow.GetUid()
				if len(addrMemoFollow.FollowAddr) == 0 {
					c.BadFollows++
					if !c.Delete {
						continue
					}
					var addrMemoFollowed = &memo.AddrHeightFollowed{
						FollowAddr: addrMemoFollow.FollowAddr,
						Height:     addrMemoFollow.Height,
						TxHash:     addrMemoFollow.TxHash,
						Addr:       addrMemoFollow.Addr,
						Unfollow:   addrMemoFollow.Unfollow,
					}
					if err := db.Remove([]db.Object{addrMemoFollow, addrMemoFollowed}); err != nil {
						return jerr.Get("error removing addr memo follow/followed", err)
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
