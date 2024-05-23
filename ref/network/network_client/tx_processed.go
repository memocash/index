package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"time"
)

type TxProcessed struct {
	TxHash    []byte
	Timestamp time.Time
}

func (t *TxProcessed) Get(txHash []byte) error {
	conn, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer conn.Close()
	response, err := conn.Client.ListenTx(conn.GetDefaultContext(), &network_pb.TxRequest{
		Hash: txHash,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network listen tx; %w", err)
	}
	t.Timestamp = time.Unix(response.Timestamp, 0)
	return nil
}

func NewTxProcessed() *TxProcessed {
	return &TxProcessed{
	}
}
