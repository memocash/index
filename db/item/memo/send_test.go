package memo

import "testing"

func TestSendRoundTrip(t *testing.T) {
	send := &Send{
		TxHash:    [32]byte{1, 2, 3},
		Recipient: [25]byte{4, 5, 6},
	}
	var decoded Send
	decoded.SetUid(send.GetUid())
	decoded.Deserialize(send.Serialize())
	if decoded != *send {
		t.Fatalf("decoded send = %+v, want %+v", decoded, *send)
	}
}

func TestSendRejectsInvalidLengths(t *testing.T) {
	var send Send
	send.SetUid([]byte{1})
	send.Deserialize([]byte{2})
	if send != (Send{}) {
		t.Fatalf("send changed after invalid data: %+v", send)
	}
}
