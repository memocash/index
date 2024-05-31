package attach

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoLike struct {
	base
	Likes       []*model.Like
	DetailsWait sync.WaitGroup
}

func ToMemoLikes(ctx context.Context, fields []Field, likes []*model.Like) error {
	if len(likes) == 0 {
		return nil
	}
	o := MemoLike{
		base:  base{Ctx: ctx, Fields: fields},
		Likes: likes,
	}
	o.DetailsWait.Add(1)
	go o.AttachInfo()
	o.Wait.Add(4)
	go o.AttachTips()
	go o.AttachTxs()
	go o.DetailsWait.Wait()
	go o.AttachLocks()
	go o.AttachPosts()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo likes; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLike) getTxHashes(checkAddressPostTxHash, checkTips bool) [][32]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var txHashes [][32]byte
	for i := range a.Likes {
		if checkAddressPostTxHash &&
			!jutil.AllZeros(a.Likes[i].PostTxHash[:]) &&
			!jutil.AllZeros(a.Likes[i].Address[:]) {
			continue
		} else if checkTips && a.Likes[i].Tip > 0 {
			continue
		}
		txHashes = append(txHashes, a.Likes[i].TxHash)
	}
	return txHashes
}

func (a *MemoLike) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"address", "lock", "post_tx_hash", "post"}) {
		return
	}
	needsInfoTxHashes := a.getTxHashes(true, false)
	if len(needsInfoTxHashes) > 0 {
		a.AddError(fmt.Errorf("info needed for memo likes and this is not implemented"))
		return
	}
}

func (a *MemoLike) AttachTips() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tip"}) {
		return
	}
	memoLikeTips, err := memo.GetLikeTips(a.getTxHashes(false, true))
	if err != nil {
		a.AddError(fmt.Errorf("error getting memo like tips for post resolver; %w", err))
		return
	}
	a.Mutex.Lock()
	for _, memoLikeTip := range memoLikeTips {
		for _, like := range a.Likes {
			if memoLikeTip.LikeTxHash == like.PostTxHash {
				like.Tip = memoLikeTip.Tip
			}
		}
	}
	a.Mutex.Unlock()
}

func (a *MemoLike) AttachLocks() {
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
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo likes; %w", err))
		return
	}
}

func (a *MemoLike) AttachTxs() {
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
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo likes; %w", err))
		return
	}
}

func (a *MemoLike) AttachPosts() {
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
	if err := ToMemoPosts(a.Ctx, GetPrefixFields(a.Fields, "post."), allPosts); err != nil {
		a.AddError(fmt.Errorf("error attaching to posts for memo likes; %w", err))
		return
	}
}
