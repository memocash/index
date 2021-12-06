package network_client

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type BlockTx struct {
	BlockHash []byte
	TxHash    []byte
	Raw       []byte
}

type GetBlockTxs struct {
	BlockHash []byte
	Txs       []BlockTx
}

func (t *GetBlockTxs) GetByHeight(height int64) error {
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
	blockInfo, err := c.GetBlockByHeight(ctx, &network_pb.BlockRequest{
		Height: height,
	})
	if err != nil {
		return jerr.Get("error getting rpc network block infos by height", err)
	}
	if err := t.Get(blockInfo.Hash, nil); err != nil {
		return jerr.Get("error getting block by hash for block txs", err)
	}
	return nil
}

func (t *GetBlockTxs) Get(blockHash []byte, startTx []byte) error {
	conn, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer conn.Close()
	response, err := conn.Client.GetBlockTxs(conn.GetDefaultContext(), &network_pb.BlockTxRequest{
		Block: blockHash,
		Start: startTx,
	})
	if err != nil {
		return jerr.Get("error getting rpc network block infos", err)
	}
	t.BlockHash = blockHash
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

func NewGetBlockTxs() *GetBlockTxs {
	return &GetBlockTxs{
	}
}
