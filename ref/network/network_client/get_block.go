package network_client

import (
	"context"
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type GetBlock struct {
	Block *BlockInfo
}

func (b *GetBlock) GetByHeight(height int64) error {
	rpcConfig := GetConfig()
	if ! rpcConfig.IsSet() {
		return fmt.Errorf("error config not set")
	}
	conn, err := grpc.Dial(rpcConfig.String(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("error dial grpc did not connect network; %w", err)
	}
	defer conn.Close()
	c := network_pb.NewNetworkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	blockInfo, err := c.GetBlockByHeight(ctx, &network_pb.BlockRequest{
		Height: height,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network block infos by height; %w", err)
	}
	b.Block = &BlockInfo{
		Hash:   blockInfo.Hash,
		Height: blockInfo.Height,
		Txs:    blockInfo.Txs,
		Header: blockInfo.Header,
	}
	return nil
}

func (b *GetBlock) GetByHash(hash []byte) error {
	rpcConfig := GetConfig()
	if ! rpcConfig.IsSet() {
		return fmt.Errorf("error config not set")
	}
	conn, err := grpc.Dial(rpcConfig.String(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("error dial grpc did not connect network; %w", err)
	}
	defer conn.Close()
	c := network_pb.NewNetworkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	blockInfo, err := c.GetBlockByHash(ctx, &network_pb.BlockHashRequest{
		Hash: hash,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network block infos by hash; %w", err)
	}
	b.Block = &BlockInfo{
		Hash:   blockInfo.Hash,
		Height: blockInfo.Height,
		Txs:    blockInfo.Txs,
		Header: blockInfo.Header,
	}
	return nil
}

func NewGetBlock() *GetBlock {
	return &GetBlock{
	}
}
