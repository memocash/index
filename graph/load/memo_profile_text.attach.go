package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"sync"
)

type MemoSetProfileAttach struct {
	baseA
	SetProfiles []*model.SetProfile
	DetailsWait sync.WaitGroup
}

func AttachToMemoSetProfiles(ctx context.Context, fields []Field, setProfiles []*model.SetProfile) error {
	if len(setProfiles) == 0 {
		return nil
	}
	o := MemoSetProfileAttach{
		baseA:       baseA{Ctx: ctx, Fields: fields},
		SetProfiles: setProfiles,
	}
	o.DetailsWait.Add(1)
	go o.AttachInfo()
	o.Wait.Add(2)
	go o.AttachLocks()
	o.DetailsWait.Wait()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo set profiles; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoSetProfileAttach) getAddresses() [][25]byte {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	var addresses [][25]byte
	for i := range a.SetProfiles {
		addresses = append(addresses, a.SetProfiles[i].Address)
	}
	return addresses
}

func (a *MemoSetProfileAttach) AttachInfo() {
	defer a.DetailsWait.Done()
	if !a.HasField([]string{"tx", "tx_hash", "profile"}) {
		return
	}
	addrProfiles, err := memo.GetAddrProfiles(a.Ctx, a.getAddresses())
	if err != nil {
		a.AddError(fmt.Errorf("error getting addr profile profiles for set profile attach; %w", err))
		return
	}
	a.Mutex.Lock()
	for _, addrProfile := range addrProfiles {
		for _, setProfile := range a.SetProfiles {
			if addrProfile.Addr == setProfile.Address {
				setProfile.TxHash = addrProfile.TxHash
				setProfile.Text = addrProfile.Profile
			}
		}
	}
	a.Mutex.Unlock()
}

func (a *MemoSetProfileAttach) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, setProfile := range a.SetProfiles {
		setProfile.Lock = &model.Lock{Address: setProfile.Address}
		allLocks = append(allLocks, setProfile.Lock)
	}
	a.Mutex.Unlock()
	if err := AttachToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo set profiles; %w", err))
		return
	}
}

func (a *MemoSetProfileAttach) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, setProfile := range a.SetProfiles {
		setProfile.Tx = &model.Tx{Hash: setProfile.TxHash}
		allTxs = append(allTxs, setProfile.Tx)
	}
	a.Mutex.Unlock()
	if err := AttachToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo set profiles; %w", err))
		return
	}
}
