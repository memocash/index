package network_client

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
	"google.golang.org/grpc"
	"time"
)

type Balance struct {
	Address   string
	Balance   int64
	Spendable int64
	UtxoCount int
	Spends    int
	Outputs   int
	UTXOs     int
	Txs       int
}

func (b *Balance) GetByAddress(address string) error {
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
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	balance, err := c.GetBalance(ctx, &network_pb.Address{
		Address: address,
	})
	if err != nil {
		return jerr.Get("error getting rpc network balance by address", err)
	}
	b.Address = balance.Address
	b.Balance = balance.Balance
	b.Spendable = balance.Spendable
	b.Spends = int(balance.Spends)
	b.UtxoCount = int(balance.Utxos)
	return nil
}

func (b *Balance) GetByAddresses(addresses []string) error {
	var totalBalance int64
	var totalSpendable int64
	var totalUtxoCount int
	var totalSpends int
	for _, address := range addresses {
		if err := b.GetByAddress(address); err != nil {
			return jerr.Getf(err, "error getting balance for address: %s", address)
		}
		totalBalance += b.Balance
		totalSpendable += b.Spendable
		totalUtxoCount += b.UtxoCount
		totalSpends += b.Spends
	}
	b.Balance = totalBalance
	b.Spendable = totalSpendable
	b.UtxoCount = totalUtxoCount
	b.Spends = totalSpends
	return nil
}

func NewBalance() *Balance {
	return &Balance{
	}
}
