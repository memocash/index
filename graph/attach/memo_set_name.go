package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoSetName struct {
	base
	SetNames    []*model.SetName
	DetailsWait sync.WaitGroup
}

func ToMemoSetNames(ctx context.Context, fields []Field, setNames []*model.SetName) error {
	if len(setNames) == 0 {
		return nil
	}
	o := MemoSetName{
		base:     base{Ctx: ctx, Fields: fields},
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

func (a *MemoSetName) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	for i := range a.SetNames {
		addresses = append(addresses, a.SetNames[i].Address)
	}
	return addresses
}

func (a *MemoSetName) getSetNameIndexMap() map[[25]byte][]int {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	m := make(map[[25]byte][]int, len(a.SetNames))
	for i := range a.SetNames {
		m[a.SetNames[i].Address] = append(m[a.SetNames[i].Address], i)
	}
	return m
}

func (a *MemoSetName) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"tx", "tx_hash", "name"}) {
		return
	}
	addrProfileNames, err := memo.GetAddrNames(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profile names for set name attach; %w", err))
		return
	}
	setNameIndexMap := a.getSetNameIndexMap()
	a.Mutex.Lock()
	for _, addrProfileName := range addrProfileNames {
		for _, i := range setNameIndexMap[addrProfileName.Addr] {
			a.SetNames[i].TxHash = addrProfileName.TxHash
			a.SetNames[i].Name = addrProfileName.Name
		}
	}
	a.Mutex.Unlock()
}

func (a *MemoSetName) AttachLocks() {
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
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo set names; %w", err))
		return
	}
}

func (a *MemoSetName) AttachTxs() {
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
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo set names; %w", err))
		return
	}
}
