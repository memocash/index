package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	driver "github.com/memocash/index/client/drivers/sql"
	"github.com/memocash/index/client/lib"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	if len(os.Args) < 2 {
		log.Fatal("error no command provided")
	}
	switch os.Args[1] {
	case "utxos":
		if len(os.Args) < 3 {
			log.Fatal("no address provided")
		}
		var addresses []wallet.Addr
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "verbose" {
				continue
			}
			address, err := wallet.GetAddrFromString(os.Args[i])
			if err != nil {
				log.Fatalf("error getting address from string; %v", err)
			}
			addresses = append(addresses, *address)
		}
		client, err := GetClient()
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
	case "balance":
		if len(os.Args) < 3 {
			log.Fatal("no address provided")
		}
		var addresses []wallet.Addr
		for i := 2; i < len(os.Args); i++ {
			address, err := wallet.GetAddrFromString(os.Args[i])
			if err != nil {
				log.Fatalf("error getting address from string; %v", err)
			}
			addresses = append(addresses, *address)
		}
		client, err := GetClient()
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
}

func GetClient() (*lib.Client, error) {
	db, err := sql.Open("sqlite3", "./example.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database; %w", err)
	}
	database, err := driver.NewDatabase(db, "example")
	if err != nil {
		return nil, fmt.Errorf("error getting database; %w", err)
	}
	return lib.NewClient(graph.DefaultUrl, database), nil
}
