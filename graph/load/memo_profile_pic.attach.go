package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoSetPicAttach struct {
	baseA
	SetPics     []*model.SetPic
	DetailsWait sync.WaitGroup
}

func AttachToMemoSetPics(ctx context.Context, fields []Field, setPics []*model.SetPic) error {
	if len(setPics) == 0 {
		return nil
	}
	o := MemoSetPicAttach{
		baseA:   baseA{Ctx: ctx, Fields: fields},
		SetPics: setPics,
	}
	o.DetailsWait.Add(1)
	go o.AttachInfo()
	o.Wait.Add(2)
	go o.AttachLocks()
	o.DetailsWait.Wait()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo set pics; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoSetPicAttach) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	for i := range a.SetPics {
		addresses = append(addresses, a.SetPics[i].Address)
	}
	return addresses
}

func (a *MemoSetPicAttach) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"tx", "tx_hash", "pic"}) {
		return
	}
	addrProfilePics, err := memo.GetAddrProfilePics(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profile pics for set pic attach; %w", err))
		return
	}
	a.Mutex.Lock()
	for _, addrProfilePic := range addrProfilePics {
		for _, setPic := range a.SetPics {
			if addrProfilePic.Addr == setPic.Address {
				setPic.TxHash = addrProfilePic.TxHash
				setPic.Pic = addrProfilePic.Pic
			}
		}
	}
	a.Mutex.Unlock()
}

func (a *MemoSetPicAttach) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, setPic := range a.SetPics {
		setPic.Lock = &model.Lock{Address: setPic.Address}
		allLocks = append(allLocks, setPic.Lock)
	}
	a.Mutex.Unlock()
	if err := AttachToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo set pics; %w", err))
		return
	}
}

func (a *MemoSetPicAttach) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, setPic := range a.SetPics {
		setPic.Tx = &model.Tx{Hash: setPic.TxHash}
		allTxs = append(allTxs, setPic.Tx)
	}
	a.Mutex.Unlock()
	if err := AttachToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo set pics; %w", err))
		return
	}
}
