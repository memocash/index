package sub

import (
	"context"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/dataloader"
	"github.com/memocash/index/admin/graph/load"
	"github.com/memocash/index/admin/graph/model"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
)

type Post struct {
	Name         string
	PostHashChan chan []byte
	Cancel       context.CancelFunc
}

func (r *Post) Listen(ctx context.Context, txHashes []string) (<-chan *model.Post, error) {
	var txHashesBytes = make([][]byte, len(txHashes))
	for i := range txHashes {
		txHash, err := chainhash.NewHashFromStr(txHashes[i])
		if err != nil {
			return nil, jerr.Get("error getting tx hash from string for post subscription", err)
		}
		txHashesBytes[i] = txHash.CloneBytes()
	}
	ctx, r.Cancel = context.WithCancel(ctx)
	var postChan = make(chan *model.Post)
	r.PostHashChan = make(chan []byte)
	postChildListener, err := memo.ListenPostChildren(ctx, txHashesBytes)
	if err != nil {
		r.Cancel()
		return nil, jerr.Get("error getting memo post child listener for post subscription", err)
	}
	postLikesListener, err := memo.ListenPostHeightLikes(ctx, txHashesBytes)
	if err != nil {
		r.Cancel()
		return nil, jerr.Get("error getting memo post likes listener for post subscription", err)
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
				post, err := dataloader.NewPostLoader(load.PostLoaderConfig).Load(hs.GetTxString(txHash))
				if err != nil {
					jerr.Get("error getting post from dataloader for post subscription resolver", err).Print()
					return
				}
				postChan <- post
			}
		}
	}()
	return postChan, nil
}
