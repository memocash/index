package network_client

import (
	"context"
	"fmt"
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
		return fmt.Errorf("did not connect network; %w", err)
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
		return fmt.Errorf("error connection save tx block; %w", err)
	}
	if errorReply.Error != "" {
		return fmt.Errorf("error save tx block reply: %s", errorReply.Error)
	}
	return nil
}
