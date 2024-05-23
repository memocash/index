package sub

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/load"
	"github.com/memocash/index/graph/model"
	"log"
)

type Post struct {
	Name         string
	PostHashChan chan [32]byte
	Cancel       context.CancelFunc
}

func (r *Post) Listen(ctx context.Context, txHashes [][32]byte) (<-chan *model.Post, error) {
	ctx, r.Cancel = context.WithCancel(ctx)
	var postChan = make(chan *model.Post)
	r.PostHashChan = make(chan [32]byte)
	if len(txHashes) == 0 {
		postListener, err := memo.ListenPosts(ctx)
		if err != nil {
			r.Cancel()
			return nil, fmt.Errorf("error getting memo post child listener for post subscription; %w", err)
		}
		go func() {
			defer r.Cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case post, ok := <-postListener:
					if !ok {
						return
					}
					r.PostHashChan <- post.TxHash
				}
			}
		}()
	} else {
		postChildListener, err := memo.ListenPostChildren(ctx, txHashes)
		if err != nil {
			r.Cancel()
			return nil, fmt.Errorf("error getting memo post child listener for post subscription; %w", err)
		}
		postLikesListener, err := memo.ListenPostLikes(ctx, txHashes)
		if err != nil {
			r.Cancel()
			return nil, fmt.Errorf("error getting memo post likes listener for post subscription; %w", err)
		}
		go func() {
			defer r.Cancel()
			for {
				select {
				case <-ctx.Done():
					return
				case postChild, ok := <-postChildListener:
					if !ok {
						return
					}
					r.PostHashChan <- postChild.PostTxHash
				case postLike, ok := <-postLikesListener:
					if !ok {
						return
					}
					r.PostHashChan <- postLike.PostTxHash
				}
			}
		}()
	}
	go func() {
		defer func() {
			close(r.PostHashChan)
			close(postChan)
			r.Cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case txHash, ok := <-r.PostHashChan:
				if !ok {
					return
				}
				post, err := load.Post.Load(chainhash.Hash(txHash).String())
				if err != nil {
					log.Printf("error getting post from dataloader for post subscription resolver; %v", err)
					return
				}
				postChan <- post
			}
		}
	}()
	return postChan, nil
}
