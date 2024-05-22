package main

import (
	"encoding/hex"
	"example/common"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("no parent tx hash provided")
	}
	parentHash, err := chainhash.NewHashFromStr(os.Args[1])
	if err != nil {
		log.Fatalf("error parsing reply tx hash; %v", err)
	}
	var msg = strings.Join(os.Args[2:], " ")
	if msg == "" {
		log.Fatalf("No memo reply message provided.")
	}
	wlt, err := common.NewWalletFromStdinWif()
	if err != nil {
		log.Fatalf("error creating new wallet; %v", err)
	}
	tx, err := gen.Tx(gen.TxRequest{
		Getter: wlt,
		Outputs: []*memo.Output{{Script: &script.Reply{
			TxHash:  parentHash[:],
			Message: msg,
		}}},
		Change:  wlt.Change,
		KeyRing: wlt.KeyRing,
	})
	if err != nil {
		log.Fatalf("error generating memo reply tx; %v", err)
	}
	txInfo := parse.GetTxInfo(tx)
	txInfo.Print()
	if err := wlt.Client.Broadcast(hex.EncodeToString(txInfo.Raw)); err != nil {
		log.Fatalf("error broadcasting memo reply tx; %v", err)
	}
	log.Println("Memo reply tx broadcast!")
}
