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
