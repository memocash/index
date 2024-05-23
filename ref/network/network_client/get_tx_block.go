package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type GetTxBlock struct {
	Txs []BlockTx
}

func (t *GetTxBlock) Get(txHashes [][]byte) error {
	conn, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer conn.Close()
	response, err := conn.Client.GetTxBlock(conn.GetDefaultContext(), &network_pb.TxBlockRequest{
		Txs: txHashes,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network block infos; %w", err)
	}
	t.Txs = make([]BlockTx, len(response.Txs))
	for i := range response.Txs {
		t.Txs[i] = BlockTx{
			BlockHash: response.Txs[i].Block,
			TxHash:    response.Txs[i].Tx,
			Raw:       response.Txs[i].Raw,
		}
	}
	return nil
}

func NewGetTxBlock() *GetTxBlock {
	return &GetTxBlock{
	}
}
