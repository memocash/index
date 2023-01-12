package network_client

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type BlockTxs struct {
	Header []byte
	RawTxs [][]byte
}

func SaveBlockTxs(blockTxs BlockTxs) error {
	config := GetConfig()
	if !config.IsSet() {
		return nil
	}
	conn, err := grpc.Dial(config.String(), grpc.WithInsecure())
	if err != nil {
		return jerr.Get("did not connect network", err)
	}
	defer conn.Close()
	c := network_pb.NewNetworkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var txs = make([]*network_pb.Tx, len(blockTxs.RawTxs))
	for i := range blockTxs.RawTxs {
		txs[i] = &network_pb.Tx{
			Raw: blockTxs.RawTxs[i],
		}
	}
	errorReply, err := c.SaveTxBlock(ctx, &network_pb.TxBlock{
		Block: &network_pb.Block{
			Header: blockTxs.Header,
		},
		Txs: txs,
	})
	if err != nil {
		return jerr.Get("error connection save tx block", err)
	}
	if errorReply.Error != "" {
		return jerr.Newf("error save tx block reply: %s", errorReply.Error)
	}
	return nil
}
