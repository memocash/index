package main

import (
	"example/common"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("no address provided")
	}
	var addresses []wallet.Addr
	for i := 1; i < len(os.Args); i++ {
		address, err := wallet.GetAddrFromString(os.Args[i])
		if err != nil {
			log.Fatalf("error getting address from string; %v", err)
		}
		addresses = append(addresses, *address)
	}
	client, err := common.GetClient()
	if err != nil {
		log.Fatalf("error getting client; %v", err)
	}
	balance, err := client.GetBalance(addresses)
	if err != nil {
		log.Fatalf("error getting balance; %v", err)
	}
	log.Printf("Balance: %d, utxos: %d, spendable: %d, spendable_count: %d\n",
		balance.Balance, balance.UtxoCount, balance.Spendable, balance.SpendableCount)
}
