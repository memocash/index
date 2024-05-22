package main

import (
	"example/common"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("no address provided")
	}
	wlt, err := common.NewWallet(os.Args[1:])
	if err != nil {
		log.Fatalf("error creating new wallet; %v", err)
	}
	tx, err := gen.TxUnsigned(gen.TxRequest{
		Getter: wlt,
		Outputs: []*memo.Output{{Script: &script.Post{
			Message: "Hello, world!",
		}}},
		Change: wlt.Change,
	})
	if err != nil {
		log.Fatalf("error getting unsigned tx; %v", err)
	}
	parse.GetTxInfo(tx).Print()
}
