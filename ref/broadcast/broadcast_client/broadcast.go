package broadcast_client

import (
	"context"
	"errors"
	"fmt"
	"github.com/memocash/index/ref/broadcast/gen/broadcast_pb"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/network/network_client"
)

type Broadcast struct {
}

func (t *Broadcast) Broadcast(ctx context.Context, raw []byte) error {
	conn, err := NewConnection()
	if err != nil {
		if errors.Is(err, config.NotSetError) {
			network_client.SetConfig(config.RpcConfig{
				Host: config.Localhost,
				Port: config.GetServerPort(),
			})
			if err := network_client.NewSendTx().Send([][]byte{raw}); err != nil {
				return fmt.Errorf("error sending raw txs to network; %w", err)
			}
			return nil
		}
		return fmt.Errorf("error connecting to broadcast; %w", err)
	}
	defer conn.Close()
	if _, err := conn.Client.BroadcastTx(ctx, &broadcast_pb.BroadcastRequest{
		Raw: raw,
	}); err != nil {
		return fmt.Errorf("error request rpc broadcast tx; %w", err)
	}
	return nil
}

func NewBroadcast() *Broadcast {
	return &Broadcast{}
}
