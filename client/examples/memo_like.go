package main

import (
	"encoding/hex"
	"example/common"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/client/lib/graph"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("no like tx hash provided")
	}
	parentHash, err := chainhash.NewHashFromStr(os.Args[1])
	if err != nil {
		log.Fatalf("error parsing like tx hash; %v", err)
	}
	var tip int
	if len(os.Args) > 2 {
		if tip, err = strconv.Atoi(os.Args[2]); err != nil {
			log.Fatalf("error parsing tip; %v", err)
		}
	}
	wlt, err := common.NewWalletFromStdinWif()
	if err != nil {
		log.Fatalf("error creating new wallet; %v", err)
	}
	outputs := []*memo.Output{{Script: &script.Like{
		TxHash: parentHash[:],
	}}}
	if tip > 0 {
		tx, err := graph.GetTx(parentHash.String())
		if err != nil {
			log.Fatalf("error getting like tx; %v", err)
		}
		var likeAddress string
		for _, in := range tx.Inputs {
			if in.Output.Lock.Address != "" {
				if likeAddress != "" && in.Output.Lock.Address != likeAddress {
					log.Fatalf("like tx has multiple addresses; %v and %v", likeAddress, in.Output.Lock.Address)
				}
				likeAddress = in.Output.Lock.Address
			}
		}
		if likeAddress == "" {
			log.Fatalf("no adddress found in like tx")
		}
		likeAddr, err := wallet.GetAddrFromString(likeAddress)
		if err != nil {
			log.Fatalf("error getting like address; %v", err)
		}
		outputs = append(outputs, &memo.Output{
			Script: &script.P2pkh{PkHash: likeAddr.GetPkHash()},
			Amount: int64(tip),
		})
	}
	tx, err := gen.Tx(gen.TxRequest{
		Getter:  wlt,
		Outputs: outputs,
		Change:  wlt.Change,
		KeyRing: wlt.KeyRing,
	})
	if err != nil {
		log.Fatalf("error generating memo like tx; %v", err)
	}
	txInfo := parse.GetTxInfo(tx)
	txInfo.Print()
	if err := wlt.Client.Broadcast(hex.EncodeToString(txInfo.Raw)); err != nil {
		log.Fatalf("error broadcasting memo like tx; %v", err)
	}
	log.Println("Memo like tx broadcast!")
}
