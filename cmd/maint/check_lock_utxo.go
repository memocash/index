package maint

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/node/act/maint"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/spf13/cobra"
)

var checkLockUtxoCmd = &cobra.Command{
	Use:   "check-lock-utxo",
	Short: "check-lock-utxo [block_hash]",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			jerr.New("error must specify block hash").Fatal()
		}
		blockHash, err := chainhash.NewHashFromStr(args[0])
		if err != nil {
			jerr.Get("error parsing block hash", err).Fatal()
		}
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		checkLockUtxos := maint.NewCheckLockUtxo()
		jlog.Log("Starting check lock utxo for block...")
		if err := checkLockUtxos.Check(blockHash.CloneBytes()); err != nil {
			jerr.Get("error maint check lock utxo", err).Fatal()
		}
		jlog.Logf("Checked outputs: %d, spends: %d, lock utxos: %d, missing: %d\n", checkLockUtxos.CheckedOutputs,
			checkLockUtxos.FoundInputs, checkLockUtxos.FoundUtxos, len(checkLockUtxos.MissingUtxos))
		if verbose {
			jlog.Logf("Removed outs (1): %d, removed outs (2): %d\n", checkLockUtxos.OutsRemoved1, checkLockUtxos.OutsRemoved2)
			for _, missingUtxo := range checkLockUtxos.MissingUtxos {
				jlog.Logf("unspent output without lock utxo: %s:%d\n",
					hs.GetTxString(missingUtxo.TxHash), missingUtxo.Index)
			}
		}
	},
}