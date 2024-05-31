package sub

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/attach"
	"github.com/memocash/index/graph/model"
	"log"
)

type Room struct {
	Name         string
	RoomPostChan chan [32]byte
	Cancel       context.CancelFunc
}

func (r *Room) Listen(ctx context.Context, names []string) (<-chan *model.Post, error) {
	ctx, r.Cancel = context.WithCancel(ctx)
	var postChan = make(chan *model.Post)
	r.RoomPostChan = make(chan [32]byte)
	roomHeightPostsListener, err := memo.ListenRoomPosts(ctx, names)
	if err != nil {
		r.Cancel()
		return nil, fmt.Errorf("error getting memo room height post listener for room subscription; %w", err)
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
		fields := attach.GetFields(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case txHash, ok := <-r.RoomPostChan:
				if !ok {
					return
				}
				var post = &model.Post{TxHash: txHash}
				if err := attach.ToMemoPosts(ctx, fields, []*model.Post{post}); err != nil {
					log.Printf("error attaching to memo posts for room subscription resolver; %v", err)
					return
				}
				postChan <- post
			}
		}
	}()
	return postChan, nil
}
