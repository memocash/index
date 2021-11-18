package network_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/network/gen/network_pb"
	"time"
)

type TxProcessed struct {
	TxHash    []byte
	Timestamp time.Time
}

func (t *TxProcessed) Get(txHash []byte) error {
	conn, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer conn.Close()
	response, err := conn.Client.ListenTx(conn.GetDefaultContext(), &network_pb.TxRequest{
		Hash: txHash,
	})
	if err != nil {
		return jerr.Get("error getting rpc network listen tx", err)
	}
	t.Timestamp = time.Unix(response.Timestamp, 0)
	return nil
}

func NewTxProcessed() *TxProcessed {
	return &TxProcessed{
	}
}
