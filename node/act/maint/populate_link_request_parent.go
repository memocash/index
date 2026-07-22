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

type PopulateLinkRequestParent struct {
	Ctx      context.Context
	Requests int
	Parents  int
	Skipped  int
}

func NewPopulateLinkRequestParent(ctx context.Context) *PopulateLinkRequestParent {
	return &PopulateLinkRequestParent{
		Ctx: ctx,
	}
}

func (p *PopulateLinkRequestParent) Populate() error {
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		var startUid []byte
		for {
			opt := client.OptionExLargeLimit()
			if err := dbClient.GetByPrefix(p.Ctx, db.TopicMemoAddrLinkRequest, client.NewStart(startUid), opt); err != nil {
				return fmt.Errorf("error getting db memo addr link requests for populate link request parents; %w", err)
			}
			var objects []db.Object
			for _, msg := range dbClient.Messages {
				p.Requests++
				var addrLinkRequest = new(memo.AddrLinkRequest)
				db.Set(addrLinkRequest, msg)
				startUid = addrLinkRequest.GetUid()
				if jutil.AllZeros(addrLinkRequest.ParentAddr[:]) {
					p.Skipped++
					continue
				}
				objects = append(objects, &memo.AddrLinkRequestParent{
					ParentAddr: addrLinkRequest.ParentAddr,
					Seen:       addrLinkRequest.Seen,
					TxHash:     addrLinkRequest.TxHash,
					Addr:       addrLinkRequest.Addr,
					Message:    addrLinkRequest.Message,
				})
			}
			if err := db.Save(objects); err != nil {
				return fmt.Errorf("error saving link request parents; %w", err)
			}
			p.Parents += len(objects)
			if len(dbClient.Messages) < client.ExLargeLimit {
				break
			}
		}
	}
	return nil
}
