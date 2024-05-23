package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"time"
)

type SendTx struct {
}

func (t *SendTx) Send(txs [][]byte) error {
	if err := t.SendWithBlock(txs, nil); err != nil {
		return fmt.Errorf("error sending without block; %w", err)
	}
	return nil
}

func (t *SendTx) SendWithBlock(txs [][]byte, block []byte) error {
	var networkTxs = new(network_pb.Txs)
	for i := range txs {
		networkTxs.Txs = append(networkTxs.Txs, &network_pb.Tx{
			Raw:   txs[i],
			Block: block,
		})
	}
	connection, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer connection.Close()
	if reply, err := connection.Client.SaveTxs(connection.GetTimeoutContext(5*time.Second), networkTxs); err != nil {
		return fmt.Errorf("error network client save txs request; %w", err)
	} else if reply.Error != "" {
		return fmt.Errorf("send new tx rpc error received: %s", reply.Error)
	}
	return nil
}

func NewSendTx() *SendTx {
	return &SendTx{}
}

func SendNewTx(raw []byte) error {
	if err := NewSendTx().Send([][]byte{raw}); err != nil {
		return fmt.Errorf("error sending single transaction; %w", err)
	}
	return nil
}
