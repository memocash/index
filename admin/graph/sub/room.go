package sub

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

type Room struct {
	Name         string
	RoomPostChan chan []byte
	Cancel       context.CancelFunc
}

func (r *Room) Listen(ctx context.Context, names []string) (<-chan *model.Post, error) {
	ctx, r.Cancel = context.WithCancel(ctx)
	var postChan = make(chan *model.Post)
	r.RoomPostChan = make(chan []byte)
	roomHeightPostsListener, err := memo.ListenRoomPosts(ctx, names)
	if err != nil {
		r.Cancel()
		return nil, jerr.Get("error getting memo room height post listener for room subscription", err)
	}
	go func() {
		defer r.Cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case roomHeightPost, ok := <-roomHeightPostsListener:
				if !ok {
					return
				}
				r.RoomPostChan <- roomHeightPost.TxHash
			}
		}
	}()
	go func() {
		defer func() {
			close(r.RoomPostChan)
			close(postChan)
			r.Cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case txHash, ok := <-r.RoomPostChan:
				if !ok {
					return
				}
				post, err := dataloader.NewPostLoader(load.PostLoaderConfig).Load(hs.GetTxString(txHash))
				if err != nil {
					jerr.Get("error getting post from dataloader for room subscription resolver", err).Print()
					return
				}
				postChan <- post
			}
		}
	}()
	return postChan, nil
}
