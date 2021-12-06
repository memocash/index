package network_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type GetTxBlock struct {
	Txs []BlockTx
}

func (t *GetTxBlock) Get(txHashes [][]byte) error {
	conn, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer conn.Close()
	response, err := conn.Client.GetTxBlock(conn.GetDefaultContext(), &network_pb.TxBlockRequest{
		Txs: txHashes,
	})
	if err != nil {
		return jerr.Get("error getting rpc network block infos", err)
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
