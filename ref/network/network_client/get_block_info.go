package network_client

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type BlockInfo struct {
	Hash   []byte
	Height int64
	Txs    int64
	Header []byte
}

func (i BlockInfo) GetHashString() string {
	return hs.GetTxString(i.Hash)
}

func (i BlockInfo) RawString() string {
	return hex.EncodeToString(i.Header)
}

type GetBlockInfos struct {
	Blocks []BlockInfo
}

func (t GetBlockInfos) GetMaxHeight() int64 {
	var maxHeight int64
	for _, block := range t.Blocks {
		if block.Height > maxHeight {
			maxHeight = block.Height
		}
	}
	return maxHeight
}

func (t GetBlockInfos) GetMinHeight() int64 {
	var minHeight int64
	for _, block := range t.Blocks {
		if block.Height < minHeight || minHeight == 0 {
			minHeight = block.Height
		}
	}
	return minHeight
}

func (t *GetBlockInfos) Get(startHeight int64, newestFirst bool) error {
	rpcConfig := GetConfig()
	if ! rpcConfig.IsSet() {
		return jerr.New("error config not set")
	}
	conn, err := grpc.Dial(rpcConfig.String(), grpc.WithInsecure())
	if err != nil {
		return jerr.Get("error dial grpc did not connect network", err)
	}
	defer conn.Close()
	c := network_pb.NewNetworkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	reply, err := c.GetBlockInfos(ctx, &network_pb.BlockRequest{
		Height: startHeight,
		Newest: newestFirst,
	})
	if err != nil {
		return jerr.Get("error getting rpc network block infos", err)
	}
	t.Blocks = make([]BlockInfo, len(reply.Blocks))
	for i := range reply.Blocks {
		t.Blocks[i] = BlockInfo{
			Hash:   reply.Blocks[i].Hash,
			Height: reply.Blocks[i].Height,
			Txs:    reply.Blocks[i].Txs,
			Header: reply.Blocks[i].Header,
		}
	}
	return nil
}

func NewGetBlockInfos() *GetBlockInfos {
	return &GetBlockInfos{
	}
}
