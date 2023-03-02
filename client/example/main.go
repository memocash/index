package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	driver "github.com/memocash/index/client/drivers/sql"
	"github.com/memocash/index/client/lib"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("error no command provided")
	}
	switch os.Args[1] {
	case "utxos":
		if len(os.Args) < 3 {
			log.Fatal("no address provided")
		}
		address, err := wallet.GetAddrFromString(os.Args[2])
		if err != nil {
			log.Fatalf("error getting address from string; %v", err)
		}
		client, err := GetClient()
		if err != nil {
			log.Fatalf("error getting client; %v", err)
		}
		utxos, err := client.GetUtxos(*address)
		if err != nil {
			log.Fatalf("error getting utxos; %v", err)
		}
		fmt.Printf("Utxos: %d\n", len(utxos))
	case "balance":
		if len(os.Args) < 3 {
			log.Fatal("no address provided")
		}
		address, err := wallet.GetAddrFromString(os.Args[2])
		if err != nil {
			log.Fatalf("error getting address from string; %v", err)
		}
		client, err := GetClient()
		if err != nil {
			log.Fatalf("error getting client; %v", err)
		}
		balance, err := client.GetBalance(*address)
		if err != nil {
			log.Fatalf("error getting balance; %v", err)
		}
		fmt.Printf("Balance: %d, utxos: %d, spendable: %d, spendable_count: %d\n",
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
	return lib.NewClient("http://localhost:26770/graphql", database), nil
}
