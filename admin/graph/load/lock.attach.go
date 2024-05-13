package load

import (
	"context"
	"fmt"
	"github.com/memocash/index/admin/graph/model"
)

type Lock struct {
	baseA
	Locks []*model.Lock
}

func AttachToLocks(ctx context.Context, fields []Field, locks []*model.Lock) error {
	t := Lock{
		baseA: baseA{Ctx: ctx, Fields: fields},
		Locks: locks,
	}
	t.Wait.Add(1)
	go t.AttachProfiles()
	t.Wait.Wait()
	if len(t.Errors) > 0 {
		return fmt.Errorf("error attaching details to txs; %w", t.Errors[0])
	}
	return nil
}

func (l *Lock) GetLockAddrs() []string {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	var lockAddrs []string
	for _, lock := range l.Locks {
		lockAddrs = append(lockAddrs, lock.Address)
	}
	return lockAddrs
}

func (l *Lock) AttachProfiles() {
	defer l.Wait.Done()
	if !l.HasField([]string{"profile"}) {
		return
	}
	var profiles []*model.Profile
	for _, addrString := range l.GetLockAddrs() {
		profile, err := GetProfile(l.Ctx, addrString)
		if err != nil {
			l.AddError(fmt.Errorf("error getting profile from dataloader for lock resolver; %w", err))
			return
		}
		profiles = append(profiles, profile)
	}
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	for _, lock := range l.Locks {
		for _, profile := range profiles {
			if profile.Address == lock.Address {
				lock.Profile = profile
			}
		}
	}
}
