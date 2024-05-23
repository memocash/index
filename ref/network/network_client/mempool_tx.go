package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type Tx struct {
	TxHash []byte
}

type GetMempoolTxs struct {
	Txs []Tx
}

func (t *GetMempoolTxs) Get(startTx []byte) error {
	conn, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer conn.Close()
	response, err := conn.Client.GetMempoolTxs(conn.GetDefaultContext(), &network_pb.MempoolTxRequest{
		Start: startTx,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network mempool txs; %w", err)
	}
	t.Txs = make([]Tx, len(response.Txs))
	for i := range response.Txs {
		t.Txs[i] = Tx{
			TxHash: response.Txs[i].Tx,
		}
	}
	return nil
}

func NewGetMempoolTxs() *GetMempoolTxs {
	return &GetMempoolTxs{
	}
}
