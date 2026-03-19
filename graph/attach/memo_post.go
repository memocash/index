package attach

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoPost struct {
	base
	DetailsWait sync.WaitGroup
	Posts       []*model.Post
}

func ToMemoPosts(ctx context.Context, fields []Field, posts []*model.Post) error {
	if len(posts) == 0 {
		return nil
	}
	o := MemoPost{
		base:  base{Ctx: ctx, Fields: fields},
		Posts: posts,
	}
	o.DetailsWait.Add(1)
	go o.AttachInfo()
	o.Wait.Add(6)
	go o.AttachTxs()
	go o.AttachParents()
	go o.AttachLikes()
	go o.AttachReplies()
	go o.AttachRooms()
	o.DetailsWait.Wait()
	go o.AttachLocks()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo posts; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoPost) getTxHashes(checkTextAddress bool) [][32]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var txHashes [][32]byte
	for i := range a.Posts {
		if checkTextAddress &&
			a.Posts[i].Text != "" &&
			!jutil.AllZeros(a.Posts[i].Address[:]) {
			continue
		}
		txHashes = append(txHashes, a.Posts[i].TxHash)
	}
	return txHashes
}

func (a *MemoPost) getPostIndexMap() map[[32]byte][]int {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	m := make(map[[32]byte][]int, len(a.Posts))
	for i := range a.Posts {
		m[a.Posts[i].TxHash] = append(m[a.Posts[i].TxHash], i)
	}
	return m
}

func (a *MemoPost) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"address", "text", "lock"}) {
		return
	}
	memoPosts, err := memo.GetPosts(a.Ctx, a.getTxHashes(true))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo posts for attach info; %w", err))
		return
	}
	postIndexMap := a.getPostIndexMap()
	a.Mutex.Lock()
	for _, memoPost := range memoPosts {
		indices, ok := postIndexMap[memoPost.TxHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			a.Posts[i].Address = memoPost.Addr
			a.Posts[i].Text = memoPost.Post
		}
	}
	a.Mutex.Unlock()
}

func (a *MemoPost) AttachLocks() {
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
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo posts; %w", err))
		return
	}
}

func (a *MemoPost) AttachTxs() {
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
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo posts; %w", err))
		return
	}
}

func (a *MemoPost) AttachParents() {
	defer a.Wait.Done()
	if !a.HasField([]string{"parent"}) {
		return
	}
	postParents, err := memo.GetPostParents(a.Ctx, a.getTxHashes(false))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo post parents for post attach; %w", err))
		return
	}
	var postParentTxHashes = make([][32]byte, len(postParents))
	for i := range postParents {
		postParentTxHashes[i] = postParents[i].ParentTxHash
	}
	// This verifies the parent post exists before trying to attach things to it.
	verifyParentPosts, err := memo.GetPosts(a.Ctx, postParentTxHashes)
	if err != nil {
		a.AddError(fmt.Errorf("error getting memo parent posts for post attach; %w", err))
		return
	}
	postIndexMap := a.getPostIndexMap()
	verifyMap := make(map[[32]byte]*memo.Post, len(verifyParentPosts))
	for j := range verifyParentPosts {
		verifyMap[verifyParentPosts[j].TxHash] = verifyParentPosts[j]
	}
	a.Mutex.Lock()
	var allPosts []*model.Post
	for _, postParent := range postParents {
		verifyParentPost, ok := verifyMap[postParent.ParentTxHash]
		if !ok {
			continue
		}
		indices, ok := postIndexMap[postParent.PostTxHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			a.Posts[i].Parent = &model.Post{
				TxHash: verifyParentPost.TxHash,
				Text:   verifyParentPost.Post,
			}
			allPosts = append(allPosts, a.Posts[i].Parent)
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoPosts(a.Ctx, GetPrefixFields(a.Fields, "parent."), allPosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to parents for memo posts; %w", err))
		return
	}
}

func (a *MemoPost) AttachLikes() {
	defer a.Wait.Done()
	if !a.HasField([]string{"likes"}) {
		return
	}
	memoPostLikes, err := memo.GetPostLikes(a.Ctx, a.getTxHashes(false))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo post likes for post attach; %w", err))
		return
	}
	postIndexMap := a.getPostIndexMap()
	var allLikes []*model.Like
	a.Mutex.Lock()
	for _, memoPostLike := range memoPostLikes {
		indices, ok := postIndexMap[memoPostLike.PostTxHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			like := &model.Like{
				TxHash:     memoPostLike.LikeTxHash,
				PostTxHash: memoPostLike.PostTxHash,
				Address:    memoPostLike.Addr,
			}
			a.Posts[i].Likes = append(a.Posts[i].Likes, like)
			allLikes = append(allLikes, like)
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoLikes(a.Ctx, GetPrefixFields(a.Fields, "likes."), allLikes); err != nil {
		a.AddError(fmt.Errorf("error attaching to likes for memo posts; %w", err))
		return
	}
}

func (a *MemoPost) AttachReplies() {
	defer a.Wait.Done()
	if !a.HasField([]string{"replies"}) {
		return
	}
	memoPostsChildren, err := memo.GetPostsChildren(a.Ctx, a.getTxHashes(false))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo post replies for post attach; %w", err))
		return
	}
	postIndexMap := a.getPostIndexMap()
	var allReplies []*model.Post
	a.Mutex.Lock()
	for _, memoPostChild := range memoPostsChildren {
		indices, ok := postIndexMap[memoPostChild.PostTxHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			reply := &model.Post{
				TxHash: memoPostChild.ChildTxHash,
				Parent: a.Posts[i],
			}
			a.Posts[i].Replies = append(a.Posts[i].Replies, reply)
			allReplies = append(allReplies, reply)
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoPosts(a.Ctx, GetPrefixFields(a.Fields, "replies."), allReplies); err != nil {
		a.AddError(fmt.Errorf("error attaching to replies for memo posts; %w", err))
		return
	}
}

func (a *MemoPost) AttachRooms() {
	defer a.Wait.Done()
	if !a.HasField([]string{"room"}) {
		return
	}
	postRooms, err := memo.GetPostRooms(a.Ctx, a.getTxHashes(false))
	if err != nil && !client.IsEntryNotFoundError(err) {
		a.AddError(fmt.Errorf("error getting memo post rooms for post attach; %w", err))
		return
	}
	postIndexMap := a.getPostIndexMap()
	var allRooms []*model.Room
	a.Mutex.Lock()
	for _, postRoom := range postRooms {
		indices, ok := postIndexMap[postRoom.TxHash]
		if !ok {
			continue
		}
		for _, i := range indices {
			a.Posts[i].Room = &model.Room{Name: postRoom.Room}
			allRooms = append(allRooms, a.Posts[i].Room)
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoRooms(a.Ctx, GetPrefixFields(a.Fields, "room."), allRooms); err != nil {
		a.AddError(fmt.Errorf("error attaching to rooms for memo posts; %w", err))
		return
	}
}
