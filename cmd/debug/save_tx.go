package debug

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
	"log"
)

var saveTxCmd = &cobra.Command{
	Use:   "save-tx",
	Short: "save-tx [raw]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("not enough arguments, must specify raw tx")
		}
		txSaver := saver.NewCombinedTx(false)
		txRaw, err := hex.DecodeString(args[0])
		if err != nil {
			log.Fatalf("error decoding tx; %v", err)
		}
		tx, err := memo.GetMsgFromRaw(txRaw)
		if err != nil {
			log.Fatalf("error getting msg tx; %v", err)
		}
		txInfo := parse.GetTxInfoMsg(tx)
		txInfo.Print()
		block := dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{tx}, nil))
		if err := txSaver.SaveTxs(context.Background(), block); err != nil {
			log.Fatalf("error saving funding tx; %v", err)
		}
	},
}
