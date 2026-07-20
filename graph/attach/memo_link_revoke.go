package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/graph/model"
)

type MemoLinkRevoke struct {
	base
	LinkRevokes []*model.LinkRevoke
}

func ToMemoLinkRevokes(ctx context.Context, fields []Field, linkRevokes []*model.LinkRevoke) error {
	if len(linkRevokes) == 0 {
		return nil
	}
	o := MemoLinkRevoke{
		base:        base{Ctx: ctx, Fields: fields},
		LinkRevokes: linkRevokes,
	}
	o.Wait.Add(2)
	go o.AttachLocks()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo link revokes; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLinkRevoke) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, linkRevoke := range a.LinkRevokes {
		linkRevoke.Lock = &model.Lock{Address: linkRevoke.Address}
		allLocks = append(allLocks, linkRevoke.Lock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo link revokes; %w", err))
		return
	}
}

func (a *MemoLinkRevoke) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, linkRevoke := range a.LinkRevokes {
		linkRevoke.Tx = &model.Tx{Hash: linkRevoke.TxHash}
		allTxs = append(allTxs, linkRevoke.Tx)
	}
	a.Mutex.Unlock()
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo link revokes; %w", err))
		return
	}
}
