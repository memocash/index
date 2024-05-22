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
		if os.Args[i] == "verbose" {
			continue
		}
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
	utxos, err := client.GetUtxos(addresses)
	if err != nil {
		log.Fatalf("error getting utxos; %v", err)
	}
	log.Printf("Utxos: %d\n", len(utxos))
	if len(os.Args) >= 4 && os.Args[3] == "verbose" {
		for _, utxo := range utxos {
			log.Printf("utxo: %s:%d - %d\n", utxo.Hash, utxo.Index, utxo.Amount)
		}
	}
}
