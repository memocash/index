package memo

import (
	"bytes"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"testing"
	"time"
)

func TestGetRoomPostsPrefix(t *testing.T) {
	roomHash := GetRoomHash("memo")
	start := time.Date(2026, time.July, 21, 12, 34, 56, 789, time.UTC)
	var txHash [32]byte
	for i := range txHash {
		txHash[i] = byte(i)
	}

	prefix := getRoomPostsPrefix(roomHash, start, txHash, 25, true)
	wantStart := jutil.CombineBytes(roomHash, jutil.GetTimeByteNanoBig(start), jutil.ByteReverse(txHash[:]))
	if !bytes.Equal(prefix.Prefix, roomHash) {
		t.Fatalf("prefix = %x, want %x", prefix.Prefix, roomHash)
	}
	if !bytes.Equal(prefix.Start, wantStart) {
		t.Fatalf("start = %x, want %x", prefix.Start, wantStart)
	}
	if prefix.Limit != 25 {
		t.Fatalf("limit = %d, want 25", prefix.Limit)
	}
}

func TestGetRoomPostsPrefixExcludesAscendingCursor(t *testing.T) {
	roomHash := GetRoomHash("memo")
	start := time.Date(2026, time.July, 21, 12, 34, 56, 789, time.UTC)
	var txHash [32]byte

	prefix := getRoomPostsPrefix(roomHash, start, txHash, 25, false)
	wantStart := append(jutil.CombineBytes(roomHash, jutil.GetTimeByteNanoBig(start), jutil.ByteReverse(txHash[:])), 0)
	if !bytes.Equal(prefix.Start, wantStart) {
		t.Fatalf("start = %x, want %x", prefix.Start, wantStart)
	}
}

func TestGetRoomPostsPrefixUsesDefaultAndMaximum(t *testing.T) {
	roomHash := GetRoomHash("memo")

	tests := []struct {
		limit uint32
		want  uint32
	}{
		{limit: 0, want: DefaultPageSize},
		{limit: client.MediumLimit + 1, want: client.MediumLimit + 1},
		{limit: MaxPageSize + 1, want: MaxPageSize},
	}
	for _, test := range tests {
		prefix := getRoomPostsPrefix(roomHash, time.Time{}, [32]byte{}, test.limit, true)
		if len(prefix.Start) != 0 {
			t.Fatalf("start = %x, want empty", prefix.Start)
		}
		if prefix.Limit != test.want {
			t.Fatalf("limit = %d, want %d", prefix.Limit, test.want)
		}
	}
}
