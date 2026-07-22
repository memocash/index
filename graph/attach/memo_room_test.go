package attach

import (
	"fmt"
	"github.com/memocash/index/db/item/memo"
	"github.com/memocash/index/graph/model"
	"strings"
	"testing"
)

func TestAttachRoomPostsRejectsInvalidPagination(t *testing.T) {
	tests := []struct {
		name      string
		arguments map[string]interface{}
		rooms     []*model.Room
		wantError string
	}{
		{
			name:      "transaction without date",
			arguments: map[string]interface{}{"tx": "unused"},
			rooms:     []*model.Room{{Name: "one"}},
			wantError: "tx cursor requires start",
		},
		{
			name:      "cursor across rooms",
			arguments: map[string]interface{}{"start": "2026-07-22T12:00:00Z"},
			rooms:     []*model.Room{{Name: "one"}, {Name: "two"}},
			wantError: "cursor cannot be used with multiple rooms",
		},
		{
			name:      "limit above maximum",
			arguments: map[string]interface{}{"limit": fmt.Sprint(memo.MaxPageSize + 1)},
			rooms:     []*model.Room{{Name: "one"}},
			wantError: "exceeds maximum",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := MemoRoom{
				base:  base{Fields: Fields{{Name: "posts", Arguments: test.arguments}}},
				Rooms: test.rooms,
			}
			a.Wait.Add(1)
			a.AttachPosts()
			if len(a.Errors) != 1 || !strings.Contains(a.Errors[0].Error(), test.wantError) {
				t.Fatalf("AttachPosts() errors = %v, want %q", a.Errors, test.wantError)
			}
		})
	}
}

func TestAttachRoomFollowersRejectsTransactionWithoutDate(t *testing.T) {
	a := MemoRoom{
		base: base{Fields: Fields{{
			Name:      "followers",
			Arguments: map[string]interface{}{"tx": "unused"},
		}}},
		Rooms: []*model.Room{{Name: "one"}},
	}
	a.Wait.Add(1)
	a.AttachFollowers()
	if len(a.Errors) != 1 || !strings.Contains(a.Errors[0].Error(), "tx cursor requires start") {
		t.Fatalf("AttachFollowers() errors = %v, want tx cursor error", a.Errors)
	}
}

func TestAttachRoomFollowersRejectsLimitAboveMaximum(t *testing.T) {
	a := MemoRoom{
		base: base{Fields: Fields{{
			Name:      "followers",
			Arguments: map[string]interface{}{"limit": fmt.Sprint(memo.MaxPageSize + 1)},
		}}},
		Rooms: []*model.Room{{Name: "one"}},
	}
	a.Wait.Add(1)
	a.AttachFollowers()
	if len(a.Errors) != 1 || !strings.Contains(a.Errors[0].Error(), "exceeds maximum") {
		t.Fatalf("AttachFollowers() errors = %v, want maximum error", a.Errors)
	}
}
