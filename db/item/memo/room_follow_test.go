package memo

import (
	"bytes"
	"github.com/jchavannes/jgo/jutil"
	"testing"
	"time"
)

func TestRoomFollowSerializeDeserialize(t *testing.T) {
	var address [25]byte
	for i := range address {
		address[i] = byte(i + 1)
	}

	original := RoomFollow{Addr: address, Unfollow: true}
	serialized := original.Serialize()
	if len(serialized) != 26 {
		t.Fatalf("serialized length = %d, want 26", len(serialized))
	}

	var decoded RoomFollow
	decoded.Deserialize(serialized)
	if decoded.Addr != original.Addr {
		t.Fatalf("decoded address = %x, want %x", decoded.Addr, original.Addr)
	}
	if decoded.Unfollow != original.Unfollow {
		t.Fatalf("decoded unfollow = %t, want %t", decoded.Unfollow, original.Unfollow)
	}
}

func TestRoomFollowDeserializeLengthBoundary(t *testing.T) {
	valid := make([]byte, 26)
	valid[0] = 1
	valid[25] = 42
	var decoded RoomFollow
	decoded.Deserialize(valid)
	if !decoded.Unfollow || decoded.Addr[24] != 42 {
		t.Fatalf("exact-length payload was not decoded: %+v", decoded)
	}

	var short RoomFollow
	short.Deserialize(valid[:25])
	if short.Unfollow || short.Addr != [25]byte{} {
		t.Fatalf("short payload changed receiver: %+v", short)
	}
}

func TestGetRoomFollowsPrefix(t *testing.T) {
	roomHash := GetRoomHash("memo")
	start := time.Date(2026, time.July, 22, 12, 34, 56, 789, time.UTC)
	var txHash [32]byte
	for i := range txHash {
		txHash[i] = byte(i)
	}

	prefix := getRoomFollowsPrefix(roomHash, start, txHash, 25)
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

func TestGetRoomFollowsPrefixUsesDefaultAndMaximum(t *testing.T) {
	roomHash := GetRoomHash("memo")

	tests := []struct {
		limit uint32
		want  uint32
	}{
		{limit: 0, want: DefaultPageSize},
		{limit: DefaultPageSize + 1, want: DefaultPageSize + 1},
		{limit: MaxPageSize + 1, want: MaxPageSize},
	}
	for _, test := range tests {
		prefix := getRoomFollowsPrefix(roomHash, time.Time{}, [32]byte{}, test.limit)
		if len(prefix.Start) != 0 {
			t.Fatalf("start = %x, want empty", prefix.Start)
		}
		if prefix.Limit != test.want {
			t.Fatalf("limit = %d, want %d", prefix.Limit, test.want)
		}
	}
}
