package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoSetNameAttach struct {
	baseA
	SetNames    []*model.SetName
	DetailsWait sync.WaitGroup
}

func AttachToMemoSetNames(ctx context.Context, fields []Field, setNames []*model.SetName) error {
	if len(setNames) == 0 {
		return nil
	}
	o := MemoSetNameAttach{
		baseA:    baseA{Ctx: ctx, Fields: fields},
		SetNames: setNames,
	}
	o.DetailsWait.Add(1)
	go o.AttachInfo()
	o.Wait.Add(2)
	go o.AttachLocks()
	o.DetailsWait.Wait()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo set names; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoSetNameAttach) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	for i := range a.SetNames {
		addresses = append(addresses, a.SetNames[i].Address)
	}
	return addresses
}

func (a *MemoSetNameAttach) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"tx", "tx_hash", "name"}) {
		return
	}
	addrProfileNames, err := memo.GetAddrNames(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profile names for set name attach; %w", err))
		return
	}
	a.Mutex.Lock()
	for _, addrProfileName := range addrProfileNames {
		for _, setName := range a.SetNames {
			if addrProfileName.Name == setName.Name {
				setName.TxHash = addrProfileName.TxHash
				setName.Name = addrProfileName.Name
			}
		}
	}
	a.Mutex.Unlock()
}

func (a *MemoSetNameAttach) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, setName := range a.SetNames {
		setName.Lock = &model.Lock{Address: setName.Address}
		allLocks = append(allLocks, setName.Lock)
	}
	a.Mutex.Unlock()
	if err := AttachToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo set names; %w", err))
		return
	}
}

func (a *MemoSetNameAttach) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, setName := range a.SetNames {
		setName.Tx = &model.Tx{Hash: setName.TxHash}
		allTxs = append(allTxs, setName.Tx)
	}
	a.Mutex.Unlock()
	if err := AttachToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo set names; %w", err))
		return
	}
}
