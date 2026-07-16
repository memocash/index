package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/graph/model"
)

type MemoLinkAccept struct {
	base
	LinkAccepts []*model.LinkAccept
}

func ToMemoLinkAccepts(ctx context.Context, fields []Field, linkAccepts []*model.LinkAccept) error {
	if len(linkAccepts) == 0 {
		return nil
	}
	o := MemoLinkAccept{
		base:        base{Ctx: ctx, Fields: fields},
		LinkAccepts: linkAccepts,
	}
	o.Wait.Add(2)
	go o.AttachLocks()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo link accepts; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLinkAccept) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, linkAccept := range a.LinkAccepts {
		linkAccept.Lock = &model.Lock{Address: linkAccept.Address}
		allLocks = append(allLocks, linkAccept.Lock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo link accepts; %w", err))
		return
	}
}

func (a *MemoLinkAccept) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, linkAccept := range a.LinkAccepts {
		linkAccept.Tx = &model.Tx{Hash: linkAccept.TxHash}
		allTxs = append(allTxs, linkAccept.Tx)
	}
	a.Mutex.Unlock()
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo link accepts; %w", err))
		return
	}
}
