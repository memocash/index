package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type Out struct {
	TxHash []byte
	Index  uint32
	Value  int64
	Script []byte
}

type In struct {
	TxHash    []byte
	Index     uint32
	Script    []byte
	PrevHash  []byte
	PrevIndex uint32
}

type GetOutputInput struct {
	Inputs []In
}

func (g *GetOutputInput) Get(outs []Out) error {
	conn, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer conn.Close()
	var networkOutputs = make([]*network_pb.TxHashIndex, len(outs))
	for i := range outs {
		networkOutputs[i] = &network_pb.TxHashIndex{
			Tx:    outs[i].TxHash,
			Index: outs[i].Index,
		}
	}
	response, err := conn.Client.GetOutputInputs(conn.GetDefaultContext(), &network_pb.OutputInputsRequest{
		Outputs: networkOutputs,
	})
	if err != nil {
		return fmt.Errorf("error getting rpc network output inputs; %w", err)
	}
	g.Inputs = make([]In, len(response.Inputs))
	for i := range response.Inputs {
		g.Inputs[i] = In{
			TxHash:    response.Inputs[i].Tx,
			Index:     response.Inputs[i].Index,
			Script:    response.Inputs[i].Script,
			PrevHash:  response.Inputs[i].PrevTxHash,
			PrevIndex: response.Inputs[i].PrevTxIndex,
		}
	}
	return nil
}

func NewGetOutputInput() *GetOutputInput {
	return &GetOutputInput{}
}
