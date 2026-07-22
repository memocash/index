package attach

import (
	"context"
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
)

type MemoLinkAccept struct {
	base
	LinkAccepts      []*model.LinkAccept
	RequestAddresses map[*model.LinkAccept]model.Address
}

type linkTxAddressKey struct {
	hash [32]byte
	addr [25]byte
}

var getAddrLinkRevokes = memo.GetAddrLinkRevokes

func ToMemoLinkAccepts(ctx context.Context, fields []Field, linkAccepts []*model.LinkAccept, requestAddresses map[*model.LinkAccept]model.Address) error {
	if len(linkAccepts) == 0 {
		return nil
	}
	o := MemoLinkAccept{
		base:             base{Ctx: ctx, Fields: fields},
		LinkAccepts:      linkAccepts,
		RequestAddresses: requestAddresses,
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
	addresses := make([][25]byte, 0, len(a.LinkAccepts)*2)
	requestAddresses := make([]model.Address, 0, len(a.LinkAccepts))
	seenAddresses := make(map[[25]byte]struct{}, len(a.LinkAccepts)*2)
	for _, linkAccept := range a.LinkAccepts {
		requestAddress, ok := a.RequestAddresses[linkAccept]
		if !ok {
			a.AddError(fmt.Errorf("missing request address for memo link accept %x", linkAccept.TxHash))
			return
		}
		requestAddresses = append(requestAddresses, requestAddress)
		for _, address := range [][25]byte{linkAccept.Address, requestAddress} {
			if _, ok := seenAddresses[address]; ok {
				continue
			}
			seenAddresses[address] = struct{}{}
			addresses = append(addresses, address)
		}
	}
	addrRevokes, err := getAddrLinkRevokes(a.Ctx, addresses)
	if err != nil {
		a.AddError(fmt.Errorf("error getting revokes for memo link accepts; %w", err))
		return
	}
	revokesByAccept := make(map[linkTxAddressKey][]*memo.AddrLinkRevoke)
	for _, addrRevoke := range addrRevokes {
		key := linkTxAddressKey{hash: addrRevoke.AcceptTxHash, addr: addrRevoke.Addr}
		revokesByAccept[key] = append(revokesByAccept[key], addrRevoke)
	}

	var allRevokes []*model.LinkRevoke
	a.Mutex.Lock()
	for i, linkAccept := range a.LinkAccepts {
		requestAddress := requestAddresses[i]
		// Either party to a link accept may revoke it.
		keys := [...]linkTxAddressKey{
			{hash: linkAccept.TxHash, addr: linkAccept.Address},
			{hash: linkAccept.TxHash, addr: requestAddress},
		}
		keyCount := len(keys)
		if linkAccept.Address == requestAddress {
			keyCount = 1
		}
		for _, key := range keys[:keyCount] {
			for _, addrRevoke := range revokesByAccept[key] {
				revoke := &model.LinkRevoke{
					TxHash:       addrRevoke.TxHash,
					Address:      addrRevoke.Addr,
					AcceptTxHash: addrRevoke.AcceptTxHash,
					Message:      addrRevoke.Message,
				}
				linkAccept.Revokes = append(linkAccept.Revokes, revoke)
				allRevokes = append(allRevokes, revoke)
			}
		}
	}
	a.Mutex.Unlock()
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
