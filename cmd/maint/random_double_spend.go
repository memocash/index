package maint

import (
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var randomDoubleSpendCmd = &cobra.Command{
	Use:   "random-double-spend",
	Short: "Find a random double spend in output_inputs topic",
	Run: func(c *cobra.Command, args []string) {
		r := new(maint.RandomDoubleSpend)
		if err := r.Find(); err != nil {
			log.Fatalf("error finding random double spend; %v", err)
		}
		for _, doubleSpend := range r.DoubleSpends {
			for _, spend := range doubleSpend.Spends {
				log.Printf("Found double spend: %s:%d (spending: %s:%d)\n",
					chainhash.Hash(spend.Hash), spend.Index,
					chainhash.Hash(spend.PrevHash), spend.PrevIndex)
			}
		}
	},
}
