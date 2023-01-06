package main

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/client/example/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		jerr.New("error no command provided").Fatal()
	}
	switch os.Args[1] {
	case "utxos":
		if len(os.Args) < 3 {
			jerr.New("no address provided").Fatal()
		}
		address, err := wallet.GetAddrFromString(os.Args[2])
		if err != nil {
			jerr.Get("error getting address from string", err).Fatal()
		}
		client, err := db.GetClient()
		if err != nil {
			jerr.Get("error getting client", err).Fatal()
		}
		utxos,err := client.GetUtxos(address)
		if err != nil {
			jerr.Get("error getting utxos", err).Fatal()
		}
		fmt.Printf("Utxos: %d\n", len(utxos))
	case "balance":
		if len(os.Args) < 3 {
			jerr.New("no address provided").Fatal()
		}
		address, err := wallet.GetAddrFromString(os.Args[2])
		if err != nil {
			jerr.Get("error getting address from string", err).Fatal()
		}
		client, err := db.GetClient()
		if err != nil {
			jerr.Get("error getting client", err).Fatal()
		}
		balance, err := client.GetBalance(address)
		if err != nil {
			jerr.Get("error getting balance", err).Fatal()
		}
		fmt.Printf("Balance: %d\n", balance)
	}
}
