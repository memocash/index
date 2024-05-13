package test

import (
	"context"
	"encoding/hex"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
)

var saveTxCmd = &cobra.Command{
	Use:   "save-tx",
	Short: "save-tx [raw]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			jerr.New("not enough arguments, must specify raw tx").Fatal()
		}
		txSaver := saver.NewCombinedTx(false)
		txRaw, err := hex.DecodeString(args[0])
		if err != nil {
			jerr.Get("error decoding tx", err).Fatal()
		}
		tx, err := memo.GetMsgFromRaw(txRaw)
		if err != nil {
			jerr.Get("error getting msg tx", err).Fatal()
		}
		txInfo := parse.GetTxInfoMsg(tx)
		txInfo.Print()
		block := dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{tx}, nil))
		if err := txSaver.SaveTxs(context.Background(), block); err != nil {
			jerr.Get("error saving funding tx", err).Fatal()
		}
	},
}
