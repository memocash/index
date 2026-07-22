package attach

import (
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"strings"
	"testing"
	"time"
)

func TestAttachProfileFieldsRejectLimitsAboveMaximum(t *testing.T) {
	tests := []struct {
		name   string
		attach func(*MemoProfile)
	}{
		{name: "posts", attach: (*MemoProfile).AttachPosts},
		{name: "following", attach: (*MemoProfile).AttachFollowing},
		{name: "followers", attach: (*MemoProfile).AttachFollowers},
		{name: "rooms", attach: (*MemoProfile).AttachRooms},
		{name: "links", attach: (*MemoProfile).AttachLinks},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := MemoProfile{
				base: base{Fields: Fields{{
					Name:      test.name,
					Arguments: map[string]interface{}{"limit": fmt.Sprint(memo.MaxPageSize + 1)},
				}}},
				Profiles: []*model.Profile{{}},
			}
			a.Wait.Add(1)
			test.attach(&a)
			if len(a.Errors) != 1 || !strings.Contains(a.Errors[0].Error(), "exceeds maximum") {
				t.Fatalf("Attach%s() errors = %v, want maximum error", test.name, a.Errors)
			}
		})
	}
}

func TestMergeLinkRequestsNewestFirstDeduplicatesAndLimits(t *testing.T) {
	oldest := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	middle := oldest.Add(time.Hour)
	newest := middle.Add(time.Hour)
	duplicateHash := [32]byte{2}
	requests := mergeLinkRequests(
		[]*memo.AddrLinkRequest{
			{TxHash: [32]byte{1}, Seen: oldest},
			{TxHash: duplicateHash, Seen: middle},
		},
		[]*memo.AddrLinkRequestParent{
			{TxHash: duplicateHash, Seen: middle},
			{TxHash: [32]byte{3}, Seen: newest},
		},
		2,
	)
	if len(requests) != 2 {
		t.Fatalf("mergeLinkRequests() length = %d, want 2", len(requests))
	}
	if requests[0].TxHash != (model.Hash{3}) || requests[1].TxHash != model.Hash(duplicateHash) {
		t.Fatalf("mergeLinkRequests() hashes = %x, %x; want newest unique hashes", requests[0].TxHash, requests[1].TxHash)
	}
}
