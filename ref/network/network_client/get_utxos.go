package network_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type Utxo struct {
	TxHash []byte
	Index  uint32
	Value  int64
	PkHash []byte
}

type GetUtxos struct {
	Utxos []Utxo
}

func (u *GetUtxos) Get(pkHashes [][]byte) error {
	conn, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer conn.Close()
	response, err := conn.Client.GetUtxos(conn.GetDefaultContext(), &network_pb.UtxosRequest{
		PkHashes: pkHashes,
	})
	if err != nil {
		return jerr.Get("error getting rpc network get utxos", err)
	}
	u.Utxos = make([]Utxo, len(response.Outputs))
	for i := range response.Outputs {
		u.Utxos[i] = Utxo{
			TxHash: response.Outputs[i].Tx,
			Index:  response.Outputs[i].Index,
			Value:  response.Outputs[i].Value,
			PkHash: response.Outputs[i].PkHash,
		}
	}
	return nil
}

func NewGetUtxos() *GetUtxos {
	return &GetUtxos{
	}
}
