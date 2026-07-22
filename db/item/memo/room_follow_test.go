package memo

import "testing"

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
