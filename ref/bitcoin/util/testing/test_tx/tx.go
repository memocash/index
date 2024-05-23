package test_tx

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/build"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Tx struct {
	Hash    string
	Inputs  []memo.UTXO
	Outputs []memo.UTXO
}

var fundingTx = Tx{
	Hash: "f4fe0f674e28043ee6a7bacd087df10d37b32c66aceff1689c523352549e0c29",
	Inputs: []memo.UTXO{{
		Input: memo.TxInput{
			PkHash:       GetAddressPkHash("1QCBiyfwdjXDsHghBEr5U2KxUpM2BmmJVt"),
			PrevOutHash:  GetHashBytes("2aa1adcc90296be376bf88c78a9f197c7c46152c0ce1795e65ec616b9917555a"),
			PrevOutIndex: 1,
			Value:        9982363,
		},
	}},
	Outputs: []memo.UTXO{{
		Input: memo.TxInput{
			PkHash:       GetAddressPkHash("1Pzdrdoj2NC25GMWknYn18eHYuvLoZ6dpv"),
			PrevOutHash:  GetHashBytes("f4fe0f674e28043ee6a7bacd087df10d37b32c66aceff1689c523352549e0c29"),
			PrevOutIndex: 0,
			Value:        100000,
		},
	}, {
		Input: memo.TxInput{
			PkHash:       GetAddressPkHash("1QCBiyfwdjXDsHghBEr5U2KxUpM2BmmJVt"),
			PrevOutHash:  GetHashBytes("f4fe0f674e28043ee6a7bacd087df10d37b32c66aceff1689c523352549e0c29"),
			PrevOutIndex: 1,
			Value:        9882137,
		},
	}},
}

var fundingIndex uint32

func GetFundingTx(address wallet.Address, amount int64) (*memo.Tx, error) {
	fundingIndex++
	address1wlt := GetWallet(Address1key, amount+memo.FeeP2pkh1In1OutTx, fundingIndex)
	utxos, _ := address1wlt.Getter.GetUTXOs(nil)
	tx, err := build.Send(build.SendRequest{
		Wallet:  address1wlt,
		Address: address,
		Amount:  memo.GetMaxSendForUTXOs(utxos),
	})
	if err != nil {
		return nil, fmt.Errorf("error building test funding tx; %w", err)
	}
	return tx, nil
}

func SetFundingIndex(index uint32) {
	fundingIndex = index
}
