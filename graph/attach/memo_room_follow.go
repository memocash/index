package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/graph/model"
)

type MemoRoomFollow struct {
	base
	RoomFollows []*model.RoomFollow
}

func ToMemoRoomFollows(ctx context.Context, fields []Field, roomFollows []*model.RoomFollow) error {
	if len(roomFollows) == 0 {
		return nil
	}
	o := MemoRoomFollow{
		base:        base{Ctx: ctx, Fields: fields},
		RoomFollows: roomFollows,
	}
	o.Wait.Add(3)
	go o.AttachRooms()
	go o.AttachLocks()
	go o.AttachTxs()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo room follows; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoRoomFollow) AttachRooms() {
	defer a.Wait.Done()
	if !a.HasField([]string{"room"}) {
		return
	}
	var allRooms []*model.Room
	a.Mutex.Lock()
	for _, roomFollow := range a.RoomFollows {
		roomFollow.Room = &model.Room{Name: roomFollow.Name}
		allRooms = append(allRooms, roomFollow.Room)
	}
	a.Mutex.Unlock()
	if err := ToMemoRooms(a.Ctx, GetPrefixFields(a.Fields, "room."), allRooms); err != nil {
		a.AddError(fmt.Errorf("error attaching to rooms for memo room follows; %w", err))
		return
	}
}

func (a *MemoRoomFollow) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, roomFollow := range a.RoomFollows {
		roomFollow.Lock = &model.Lock{Address: roomFollow.Address}
		allLocks = append(allLocks, roomFollow.Lock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo room follows; %w", err))
		return
	}
}

func (a *MemoRoomFollow) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, roomFollow := range a.RoomFollows {
		roomFollow.Tx = &model.Tx{Hash: roomFollow.TxHash}
		allTxs = append(allTxs, roomFollow.Tx)
	}
	a.Mutex.Unlock()
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo room follows; %w", err))
		return
	}
}

// ToMemoRoomFollowForms attaches a room follow recorded under a memo key
// and returns one copy per address form it was requested by. The memo
// fields are attached once and shared; the lock is attached per copy under
// its own address, so lock.address is the form asked for.
func ToMemoRoomFollowForms(ctx context.Context, fields Fields, roomFollow *model.RoomFollow, forms []model.Address) ([]*model.RoomFollow, error) {
	if err := ToMemoRoomFollows(ctx, fields.Without("lock"), []*model.RoomFollow{roomFollow}); err != nil {
		return nil, err
	}
	var events = make([]*model.RoomFollow, len(forms))
	var locks []*model.Lock
	for i, form := range forms {
		event := *roomFollow
		event.Address = form
		if fields.HasField("lock") {
			event.Lock = &model.Lock{Address: form}
			locks = append(locks, event.Lock)
		}
		events[i] = &event
	}
	if len(locks) > 0 {
		if err := ToLocks(ctx, GetPrefixFields(fields, "lock."), locks); err != nil {
			return nil, fmt.Errorf("error attaching to locks for memo room follow forms; %w", err)
		}
	}
	return events, nil
}
