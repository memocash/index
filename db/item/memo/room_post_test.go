package memo

import (
	"bytes"
	"testing"
	"time"

	"github.com/jchavannes/jgo/jutil"
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

func TestGetRoomPostsPrefixUsesBoundedDefault(t *testing.T) {
	roomHash := GetRoomHash("memo")

	for _, limit := range []uint32{0, RoomPostsPageSize + 1} {
		prefix := getRoomPostsPrefix(roomHash, time.Time{}, [32]byte{}, limit, true)
		if len(prefix.Start) != 0 {
			t.Fatalf("start = %x, want empty", prefix.Start)
		}
		if prefix.Limit != RoomPostsPageSize {
			t.Fatalf("limit = %d, want %d", prefix.Limit, RoomPostsPageSize)
		}
	}
}
