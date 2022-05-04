package network_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"time"
)

type SendTx struct {
}

func (t *SendTx) Send(txs [][]byte) error {
	var networkTxs = new(network_pb.Txs)
	for i := range txs {
		networkTxs.Txs = append(networkTxs.Txs, &network_pb.Tx{
			Raw:   txs[i],
			Block: nil,
		})
	}
	connection, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer connection.Close()
	if reply, err := connection.Client.SaveTxs(connection.GetTimeoutContext(5*time.Second), networkTxs); err != nil {
		return jerr.Get("error network client save txs request", err)
	} else if reply.Error != "" {
		return jerr.Newf("send new tx rpc error received: %s", reply.Error)
	}
	return nil
}

func NewSendTx() *SendTx {
	return &SendTx{}
}

func SendNewTx(raw []byte) error {
	if err := NewSendTx().Send([][]byte{raw}); err != nil {
		return jerr.Get("error sending single transaction", err)
	}
	return nil
}
