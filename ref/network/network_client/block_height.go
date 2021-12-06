package network_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type BlockHeight struct {
	Block  []byte
	Height int64
}

type BlockHeightGetter struct {
	BlockHeights []*BlockHeight
}

func (h *BlockHeightGetter) Get(startHeight int64) error {
	connection, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer connection.Close()
	if reply, err := connection.Client.GetHeightBlocks(connection.GetDefaultContext(), &network_pb.BlockHeightRequest{
		Start: startHeight,
		Wait:  false,
	}); err != nil {
		return jerr.Get("could not greet network", err)
	} else {
		h.BlockHeights = make([]*BlockHeight, len(reply.Blocks))
		for i := range reply.Blocks {
			h.BlockHeights[i] = &BlockHeight{
				Block:  reply.Blocks[i].GetHash(),
				Height: reply.Blocks[i].GetHeight(),
			}
		}
	}
	return nil
}

func NewBlockHeight() *BlockHeightGetter {
	return &BlockHeightGetter{}
}
