package attach

import (
	"context"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"testing"
)

func TestAttachAcceptsMatchesRequestHashAndParentAddress(t *testing.T) {
	originalGet := getAddrLinkAccepts
	defer func() { getAddrLinkAccepts = originalGet }()

	var requestHash [32]byte
	requestHash[0] = 1
	var parentAddress, wrongAddress [25]byte
	parentAddress[0] = 2
	wrongAddress[0] = 3
	getAddrLinkAccepts = func(context.Context, [][25]byte) ([]*memo.AddrLinkAccept, error) {
		return []*memo.AddrLinkAccept{
			{TxHash: [32]byte{4}, RequestTxHash: requestHash, Addr: wrongAddress},
			{TxHash: [32]byte{5}, RequestTxHash: requestHash, Addr: parentAddress},
		}, nil
	}

	request := &model.LinkRequest{TxHash: requestHash, ParentAddress: parentAddress}
	a := MemoLinkRequest{
		base:         base{Ctx: context.Background(), Fields: Fields{{Name: "accepts"}}},
		LinkRequests: []*model.LinkRequest{request},
	}
	a.Wait.Add(1)
	a.AttachAccepts()
	if len(a.Errors) != 0 {
		t.Fatalf("AttachAccepts() errors = %v", a.Errors)
	}
	if len(request.Accepts) != 1 || request.Accepts[0].Address != parentAddress {
		t.Fatalf("accepts = %+v, want only parent-address accept", request.Accepts)
	}
}

func TestAttachRevokesMatchesAcceptHashAndEitherParty(t *testing.T) {
	originalGet := getAddrLinkRevokes
	defer func() { getAddrLinkRevokes = originalGet }()

	var acceptHash [32]byte
	acceptHash[0] = 1
	var acceptAddress, requestAddress, wrongAddress [25]byte
	acceptAddress[0] = 2
	requestAddress[0] = 3
	wrongAddress[0] = 4
	getAddrLinkRevokes = func(context.Context, [][25]byte) ([]*memo.AddrLinkRevoke, error) {
		return []*memo.AddrLinkRevoke{
			{TxHash: [32]byte{5}, AcceptTxHash: acceptHash, Addr: acceptAddress},
			{TxHash: [32]byte{6}, AcceptTxHash: acceptHash, Addr: requestAddress},
			{TxHash: [32]byte{7}, AcceptTxHash: acceptHash, Addr: wrongAddress},
		}, nil
	}

	accept := &model.LinkAccept{TxHash: acceptHash, Address: acceptAddress}
	a := MemoLinkAccept{
		base:             base{Ctx: context.Background(), Fields: Fields{{Name: "revokes"}}},
		LinkAccepts:      []*model.LinkAccept{accept},
		RequestAddresses: map[*model.LinkAccept]model.Address{accept: requestAddress},
	}
	a.Wait.Add(1)
	a.AttachRevokes()
	if len(a.Errors) != 0 {
		t.Fatalf("AttachRevokes() errors = %v", a.Errors)
	}
	if len(accept.Revokes) != 2 {
		t.Fatalf("revokes = %+v, want both party revokes", accept.Revokes)
	}
}
