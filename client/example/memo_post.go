package main

import (
	"example/common"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
	"syscall"
)

func main() {
	fmt.Printf("Enter WIF: ")
	wif, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatalf("error reading password; %v", err)
	}
	fmt.Println()
	wlt, err := common.NewWallet(string(wif))
	if err != nil {
		log.Fatalf("error creating new wallet; %v", err)
	}
	var msg = strings.Join(os.Args[1:], " ")
	if msg == "" {
		msg = "Hello, world!"
	}
	tx, err := gen.Tx(gen.TxRequest{
		Getter: wlt,
		Outputs: []*memo.Output{{Script: &script.Post{
			Message: msg,
		}}},
		Change:  wlt.Change,
		KeyRing: wlt.KeyRing,
	})
	if err != nil {
		log.Fatalf("error getting unsigned tx; %v", err)
	}
	parse.GetTxInfo(tx).Print()
}
