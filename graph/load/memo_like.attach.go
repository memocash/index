package load

import (
	"context"
	"fmt"
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
		baseA: baseA{Ctx: ctx, Fields: fields},
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
	if !a.HasField([]string{"lock"}) {
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
	a.Mutex.Lock()
	var allPosts []*model.Post
	for _, like := range a.Likes {
		like.Post = &model.Post{TxHash: like.PostTxHash}
		allPosts = append(allPosts, like.Post)
	}
	a.Mutex.Unlock()
	if err := AttachToMemoPosts(a.Ctx, GetPrefixFields(a.Fields, "post."), allPosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to posts for memo likes; %w", err))
		return
	}
}
