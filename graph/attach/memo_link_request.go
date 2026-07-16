package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/graph/model"
)

type MemoLinkRequest struct {
	base
	LinkRequests []*model.LinkRequest
}

func ToMemoLinkRequests(ctx context.Context, fields []Field, linkRequests []*model.LinkRequest) error {
	if len(linkRequests) == 0 {
		return nil
	}
	o := MemoLinkRequest{
		base:         base{Ctx: ctx, Fields: fields},
		LinkRequests: linkRequests,
	}
	o.Wait.Add(3)
	go o.AttachLocks()
	go o.AttachParentLocks()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo link requests; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLinkRequest) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		linkRequest.Lock = &model.Lock{Address: linkRequest.Address}
		allLocks = append(allLocks, linkRequest.Lock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo link requests; %w", err))
		return
	}
}

func (a *MemoLinkRequest) AttachParentLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"parent_lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		linkRequest.ParentLock = &model.Lock{Address: linkRequest.ParentAddress}
		allLocks = append(allLocks, linkRequest.ParentLock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "parent_lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to parent locks for memo link requests; %w", err))
		return
	}
}

func (a *MemoLinkRequest) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		linkRequest.Tx = &model.Tx{Hash: linkRequest.TxHash}
		allTxs = append(allTxs, linkRequest.Tx)
	}
	a.Mutex.Unlock()
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo link requests; %w", err))
		return
	}
}
