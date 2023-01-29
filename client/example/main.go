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
			log.Fatalf("%v; error getting address from string", err)
		}
		client, err := GetClient()
		if err != nil {
			log.Fatalf("%v; error getting client", err)
		}
		utxos, err := client.GetUtxos(address)
		if err != nil {
			log.Fatalf("%v; error getting utxos", err)
		}
		fmt.Printf("Utxos: %d\n", len(utxos))
	case "balance":
		if len(os.Args) < 3 {
			log.Fatal("no address provided")
		}
		address, err := wallet.GetAddrFromString(os.Args[2])
		if err != nil {
			log.Fatalf("%v; error getting address from string", err)
		}
		client, err := GetClient()
		if err != nil {
			log.Fatalf("%v; error getting client", err)
		}
		balance, err := client.GetBalance(address)
		if err != nil {
			log.Fatalf("%v; error getting balance", err)
		}
		fmt.Printf("Balance: %d\n", balance)
	}
}

func GetClient() (*lib.Client, error) {
	db, err := sql.Open("sqlite3", "./example.db")
	if err != nil {
		return nil, fmt.Errorf("%w; error opening database", err)
	}
	database, err := driver.NewDatabase(db)
	if err != nil {
		return nil, fmt.Errorf("%w; error getting database", err)
	}
	return lib.NewClient("http://localhost:26770/graphql", database), nil
}
