package fund

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/parse"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
)

var addressCmd = &cobra.Command{
	Use: "address",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 2 {
			jerr.New("not enough arguments, must specify address and amount").Fatal()
		}
		txSaver := saver.NewCombinedTx(false, false)
		address := wallet.GetAddressFromString(args[0])
		if !address.IsSet() {
			jerr.New("invalid address").Fatal()
		}
		amount := jutil.GetInt64FromString(args[1])
		if amount < memo.DustMinimumOutput {
			jerr.New("amount must be greater than dust minimum").Fatal()
		}
		fundingTx, err := test_tx.GetFundingTx(address, amount)
		if err != nil {
			jerr.Get("error getting funding tx for address cmd", err).Fatal()
		}
		txInfo := parse.GetTxInfo(fundingTx)
		txInfo.Print()
		if err := txSaver.SaveTxs(dbi.WireBlockToBlock(memo.GetBlockFromTxs([]*wire.MsgTx{fundingTx.MsgTx}, nil))); err != nil {
			jerr.Get("error saving funding tx", err).Fatal()
		}
	},
}
