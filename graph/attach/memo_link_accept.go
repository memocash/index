package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
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
	go o.AttachTxs()
	go o.AttachRevokes()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo link accepts; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLinkAccept) AttachRevokes() {
	defer a.Wait.Done()
	if !a.HasField([]string{"revokes"}) {
		return
	}
	revokesField := a.Fields.GetField("revokes")
	var allRevokes []*model.LinkRevoke
	for _, linkAccept := range a.LinkAccepts {
		addresses := [][25]byte{linkAccept.Address}
		if linkAccept.RequestAddress != linkAccept.Address {
			addresses = append(addresses, linkAccept.RequestAddress)
		}
		addrRevokes, err := memo.GetAddrLinkRevokes(a.Ctx, addresses)
		if err != nil {
			a.AddError(fmt.Errorf("error getting revokes for memo link accept; %w", err))
			return
		}
		a.Mutex.Lock()
		for _, addrRevoke := range addrRevokes {
			if addrRevoke.AcceptTxHash != linkAccept.TxHash {
				continue
			}
			revoke := &model.LinkRevoke{
				TxHash:       addrRevoke.TxHash,
				Address:      addrRevoke.Addr,
				AcceptTxHash: addrRevoke.AcceptTxHash,
				Message:      addrRevoke.Message,
			}
			linkAccept.Revokes = append(linkAccept.Revokes, revoke)
			allRevokes = append(allRevokes, revoke)
		}
		a.Mutex.Unlock()
	}
	if err := ToMemoLinkRevokes(a.Ctx, revokesField.Fields, allRevokes); err != nil {
		a.AddError(fmt.Errorf("error attaching revokes to memo link accepts; %w", err))
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
