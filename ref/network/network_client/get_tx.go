package network_client

import (
	"context"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type GetTx struct {
	Raw       []byte
	BlockHash []byte
	Msg       *wire.MsgTx
}

func (t *GetTx) Get(hash []byte) error {
	rpcConfig := GetConfig()
	if !rpcConfig.IsSet() {
		return jerr.New("error config not set")
	}
	conn, err := grpc.Dial(rpcConfig.String(), grpc.WithInsecure())
	if err != nil {
		return jerr.Get("error dial grpc did not connect network", err)
	}
	defer conn.Close()
	c := network_pb.NewNetworkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	txReply, err := c.GetTx(ctx, &network_pb.TxRequest{
		Hash: hash,
	})
	if err != nil {
		return jerr.Get("error getting rpc network tx by hash", err)
	}
	tx := txReply.GetTx()
	t.Raw = tx.GetRaw()
	t.BlockHash = tx.GetBlock()
	t.Msg, err = memo.GetMsgFromRaw(t.Raw)
	if err != nil {
		return jerr.Get("error getting wire tx from raw", err)
	}
	return nil
}

func NewGetTx() *GetTx {
	return &GetTx{
	}
}
