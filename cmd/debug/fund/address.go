package fund

import (
	"context"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
	"log"
)

var addressCmd = &cobra.Command{
	Use: "address",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatalf("not enough arguments, must specify address and amount")
		}
		txSaver := saver.NewCombinedTx(false)
		address := wallet.GetAddressFromString(args[0])
		if !address.IsSet() {
			log.Fatalf("invalid address")
		}
		amount := jutil.GetInt64FromString(args[1])
		if amount < memo.DustMinimumOutput {
			log.Fatalf("amount must be greater than dust minimum")
		}
		fundingTx, err := test_tx.GetFundingTx(address, amount)
		if err != nil {
			log.Fatalf("error getting funding tx for address cmd; %v", err)
		}
		txInfo := parse.GetTxInfo(fundingTx)
		txInfo.Print()
		if err := txSaver.SaveTxs(context.Background(), dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{fundingTx.MsgTx}, nil))); err != nil {
			log.Fatalf("error saving funding tx; %v", err)
		}
	},
}
