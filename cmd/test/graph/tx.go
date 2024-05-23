package graph

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/client/lib/graph"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var txCmd = &cobra.Command{
	Use: "tx",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("not enough arguments, must specify tx hash")
		}
		txHash, err := chainhash.NewHashFromStr(args[0])
		if err != nil {
			log.Fatalf("error parsing tx hash; %v", err)
		}
		tx, err := graph.GetTx(txHash.String())
		if err != nil {
			log.Fatalf("error getting tx; %v", err)
		}
		log.Printf("Tx: %s (seen: %s)\n", tx.Hash, tx.Seen.Format(time.RFC3339))
		for i, input := range tx.Inputs {
			log.Printf("Input %d: %s:%d %s %d\n", i,
				input.Output.Hash, input.Output.Index, input.Output.Lock.Address, input.Output.Amount)
		}
		for i, output := range tx.Outputs {
			log.Printf("Output %d: %s %d\n", i, output.Lock.Address, output.Amount)
		}
	},
}
