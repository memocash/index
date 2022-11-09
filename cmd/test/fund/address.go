package fund

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"github.com/memocash/index/test/grp"
	"github.com/spf13/cobra"
)

var addressCmd = &cobra.Command{
	Use: "address",
	Run: func(c *cobra.Command, args []string){
		s := grp.DoubleSpend{
			TxSaver:         nil,
			DelayedTxSaver:  nil,
			DelayAmount:     0,
			BlockSaver:      nil,
			FundingPkScript: nil,
			OldBlocks:       nil,
		}
		s.TxSaver = saver.NewCombined([]dbi.TxSave{
			saver.NewTxRaw(false),
			saver.NewTx(false),
			saver.NewUtxo(false),
			saver.NewLockHeight(false),
			saver.NewDoubleSpend(false),
		})
		s.BlockSaver = saver.BlockSaver(false)

		address := wallet.GetAddressFromString(args[0])
		amount := jutil.GetInt64FromString(args[1])
		fundingTx,err := test_tx.GetFundingTx(address, amount)
		if err != nil {
			jerr.Get("error getting funding tx for addressCmd", err).Fatal()
		}
		if err := s.SaveBlock([]*memo.Tx{fundingTx}); err != nil {
			jerr.Get("error saving funding tx", err).Fatal()
		}
	},
}

