package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
)

type MemoLinkRequest struct {
	base
	LinkRequests []*model.LinkRequest
}

func ToMemoLinkRequests(ctx context.Context, fields []Field, linkRequests []*model.LinkRequest) error {
	if len(linkRequests) == 0 {
		return nil
	}
	o := MemoLinkRequest{
		base:         base{Ctx: ctx, Fields: fields},
		LinkRequests: linkRequests,
	}
	o.Wait.Add(4)
	go o.AttachLocks()
	go o.AttachParentLocks()
	go o.AttachTxs()
	go o.AttachAccepts()
	o.Wait.Wait()
	if len(o.Errors) > 0 {
		return fmt.Errorf("error attaching to memo link requests; %w", o.Errors[0])
	}
	return nil
}

func (a *MemoLinkRequest) AttachAccepts() {
	defer a.Wait.Done()
	if !a.HasField([]string{"accepts"}) {
		return
	}
	acceptsField := a.Fields.GetField("accepts")
	addresses := make([][25]byte, 0, len(a.LinkRequests))
	seenAddresses := make(map[[25]byte]struct{}, len(a.LinkRequests))
	for _, linkRequest := range a.LinkRequests {
		if _, ok := seenAddresses[linkRequest.ParentAddress]; ok {
			continue
		}
		seenAddresses[linkRequest.ParentAddress] = struct{}{}
		addresses = append(addresses, linkRequest.ParentAddress)
	}
	addrAccepts, err := memo.GetAddrLinkAccepts(a.Ctx, addresses)
	if err != nil {
		a.AddError(fmt.Errorf("error getting accepts for memo link requests; %w", err))
		return
	}
	type requestKey struct {
		hash [32]byte
		addr [25]byte
	}
	acceptsByRequest := make(map[requestKey][]*memo.AddrLinkAccept)
	for _, addrAccept := range addrAccepts {
		key := requestKey{hash: addrAccept.RequestTxHash, addr: addrAccept.Addr}
		acceptsByRequest[key] = append(acceptsByRequest[key], addrAccept)
	}

	var allAccepts []*model.LinkAccept
	requestAddresses := make(map[*model.LinkAccept]model.Address)
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		// A link accept is valid only when broadcast by the request's parent.
		key := requestKey{hash: linkRequest.TxHash, addr: linkRequest.ParentAddress}
		for _, addrAccept := range acceptsByRequest[key] {
			accept := &model.LinkAccept{
				TxHash:        addrAccept.TxHash,
				Address:       addrAccept.Addr,
				RequestTxHash: addrAccept.RequestTxHash,
				Message:       addrAccept.Message,
			}
			linkRequest.Accepts = append(linkRequest.Accepts, accept)
			allAccepts = append(allAccepts, accept)
			requestAddresses[accept] = linkRequest.Address
		}
	}
	a.Mutex.Unlock()
	if err := ToMemoLinkAccepts(a.Ctx, acceptsField.Fields, allAccepts, requestAddresses); err != nil {
		a.AddError(fmt.Errorf("error attaching accepts to memo link requests; %w", err))
	}
}

func (a *MemoLinkRequest) AttachLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		linkRequest.Lock = &model.Lock{Address: linkRequest.Address}
		allLocks = append(allLocks, linkRequest.Lock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to locks for memo link requests; %w", err))
		return
	}
}

func (a *MemoLinkRequest) AttachParentLocks() {
	defer a.Wait.Done()
	if !a.HasField([]string{"parent_lock"}) {
		return
	}
	var allLocks []*model.Lock
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		linkRequest.ParentLock = &model.Lock{Address: linkRequest.ParentAddress}
		allLocks = append(allLocks, linkRequest.ParentLock)
	}
	a.Mutex.Unlock()
	if err := ToLocks(a.Ctx, GetPrefixFields(a.Fields, "parent_lock."), allLocks); err != nil {
		a.AddError(fmt.Errorf("error attaching to parent locks for memo link requests; %w", err))
		return
	}
}

func (a *MemoLinkRequest) AttachTxs() {
	defer a.Wait.Done()
	if !a.HasField([]string{"tx"}) {
		return
	}
	var allTxs []*model.Tx
	a.Mutex.Lock()
	for _, linkRequest := range a.LinkRequests {
		linkRequest.Tx = &model.Tx{Hash: linkRequest.TxHash}
		allTxs = append(allTxs, linkRequest.Tx)
	}
	a.Mutex.Unlock()
	if err := ToTxs(a.Ctx, GetPrefixFields(a.Fields, "tx."), allTxs); err != nil {
		a.AddError(fmt.Errorf("error attaching to txs for memo link requests; %w", err))
		return
	}
}
