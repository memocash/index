package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
)

type MemoLikeAttach struct {
	baseA
	Likes []*model.Like
}

func AttachToMemoLikes(ctx context.Context, fields []Field, likes []*model.Like) error {
	if len(likes) == 0 {
		return nil
	}
	o := MemoLikeAttach{
		baseA:   baseA{Ctx: ctx, Fields: fields},
		Likes: likes,
	}
	o.Wait.Add(3)
	go o.AttachLocks()
	go o.AttachTxs()
	go o.AttachPosts()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo likes; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLikeAttach) AttachLocks() {
	defer a.Wait.Done()
	var allLocks []*model.Lock
	if a.HasField([]string{"lock"}) {
		return
	}
	a.Mutex.Lock()
	for _, like := range a.Likes {
		like.Lock = &model.Lock{Address: like.Address}
		allLocks = append(allLocks, like.Lock)
	}
	a.Mutex.Unlock()
	if err := AttachToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo likes; %w", err))
		return
	}
}

func (a *MemoLikeAttach) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, like := range a.Likes {
		like.Tx = &model.Tx{Hash: like.TxHash}
		allTxs = append(allTxs, like.Tx)
	}
	a.Mutex.Unlock()
	if err := AttachToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo likes; %w", err))
		return
	}
}

func (a *MemoLikeAttach) AttachPosts() {
	defer a.Wait.Done()
	if !a.HasField([]string{"post"}) {
		return
	}
	var txHashes [][32]byte
	a.Mutex.Lock()
	for _, like := range a.Likes {
		txHashes = append(txHashes, like.TxHash)
	}
	a.Mutex.Unlock()
	memoPosts, err := memo.GetPosts(a.Ctx, txHashes)
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo posts for like attach; %w", err))
		return
	}
	a.Mutex.Lock()
	//var allPosts []*model.Post
	for _, memoPost := range memoPosts {
		for i := range a.Likes {
			if a.Likes[i].TxHash == memoPost.TxHash {
				a.Likes[i].Post = &model.Post{
					TxHash:  memoPost.TxHash,
					Address: memoPost.Addr,
					Text:    memoPost.Post,
				}
				//allPosts = append(allPosts, a.Likes[i].Post)
			}
		}
	}
	a.Mutex.Unlock()
	/*if err := AttachToPosts(a.Ctx, GetPrefixFields(a.Fields, "post."), allPosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to posts for memo likes; %w", err))
		return
	}*/
}
