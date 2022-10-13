package broadcast_client

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/broadcast/gen/broadcast_pb"
	"github.com/memocash/index/ref/config"
)

type Broadcast struct {
}

func (t *Broadcast) Broadcast(ctx context.Context, raw []byte) error {
	conn, err := NewConnection()
	if err != nil {
		if config.IsConfigNotSetError(err) {
			return nil
		}
		return jerr.Get("error connecting to broadcast", err)
	}
	defer conn.Close()
	if _, err := conn.Client.BroadcastTx(ctx, &broadcast_pb.BroadcastRequest{
		Raw: raw,
	}); err != nil {
		return jerr.Get("error request rpc broadcast tx", err)
	}
	return nil
}

func NewBroadcast() *Broadcast {
	return &Broadcast{}
}
