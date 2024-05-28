package load

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoPostAttach struct {
	baseA
	DetailsWait sync.WaitGroup
	Posts       []*model.Post
}

func AttachToMemoPosts(ctx context.Context, fields []Field, posts []*model.Post) error {
	if len(posts) == 0 {
		return nil
	}
	o := MemoPostAttach{
		baseA: baseA{Ctx: ctx, Fields: fields},
		Posts: posts,
	}
	o.DetailsWait.Add(1)
	go o.AttachInfo()
	o.Wait.Add(3)
	go o.AttachTxs()
	go o.AttachParents()
	o.DetailsWait.Wait()
	go o.AttachLocks()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo posts; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoPostAttach) getTxHashes(checkTextAddress bool) [][32]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var txHashes [][32]byte
	for i := range a.Posts {
		if checkTextAddress && a.Posts[i].Text != "" && !jutil.AllZeros(a.Posts[i].Address[:]) {
			continue
		}
		txHashes = append(txHashes, a.Posts[i].TxHash)
	}
	return txHashes
}

func (a *MemoPostAttach) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"address", "text", "lock"}) {
		return
	}
	memoPosts, err := memo.GetPosts(a.Ctx, a.getTxHashes(true))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo posts for attach info; %w", err))
		return
	}
	a.Mutex.Lock()
	for _, memoPost := range memoPosts {
		for i := range a.Posts {
			if a.Posts[i].TxHash == memoPost.TxHash {
				a.Posts[i].Address = memoPost.Addr
				a.Posts[i].Text = memoPost.Post
			}
		}

	}
	a.Mutex.Unlock()
}

func (a *MemoPostAttach) AttachLocks() {
	defer a.Wait.Done()
	var allLocks []*model.Lock
	if !a.HasField([]string{"lock"}) {
		return
	}
	a.Mutex.Lock()
	for _, post := range a.Posts {
		post.Lock = &model.Lock{Address: post.Address}
		allLocks = append(allLocks, post.Lock)
	}
	a.Mutex.Unlock()
	if err := AttachToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo posts; %w", err))
		return
	}
}

func (a *MemoPostAttach) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, post := range a.Posts {
		post.Tx = &model.Tx{Hash: post.TxHash}
		allTxs = append(allTxs, post.Tx)
	}
	a.Mutex.Unlock()
	if err := AttachToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo posts; %w", err))
		return
	}
}

func (a *MemoPostAttach) AttachParents() {
	defer a.Wait.Done()
	if !a.HasField([]string{"parent"}) {
		return
	}
	postParents, err := memo.GetPostParents(a.Ctx, a.getTxHashes(false))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo post parents for post attach; %w", err))
		return
	}
	a.Mutex.Lock()
	var allPosts []*model.Post
	for _, postParent := range postParents {
		for i := range a.Posts {
			if a.Posts[i].TxHash == postParent.PostTxHash {
				a.Posts[i].Parent = &model.Post{TxHash: postParent.ParentTxHash}
				allPosts = append(allPosts, a.Posts[i].Parent)
			}
		}
	}
	a.Mutex.Unlock()
	if err := AttachToMemoPosts(a.Ctx, GetPrefixFields(a.Fields, "parent."), allPosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to parents for memo posts; %w", err))
		return
	}
}
