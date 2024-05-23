package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type GetDoubleSpends struct {
	Txs []Tx
}

func (s *GetDoubleSpends) Get(startTx []byte) error {
	conn, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer conn.Close()
	response, err := conn.Client.GetDoubleSpends(conn.GetDefaultContext(), &network_pb.DoubleSpendRequest{
		Start: startTx,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network double spends; %w", err)
	}
	s.Txs = make([]Tx, len(response.Txs))
	for i := range response.Txs {
		s.Txs[i] = Tx{
			TxHash: response.Txs[i].Tx,
		}
	}
	return nil
}

func NewGetDoubleSpends() *GetDoubleSpends {
	return &GetDoubleSpends{
	}
}
