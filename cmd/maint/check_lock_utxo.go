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
		checkLockUtxos := maint.NewCheckLockUtxo()
		if err := checkLockUtxos.Check(blockHash.CloneBytes()); err != nil {
			jerr.Get("error maint check lock utxo", err).Fatal()
		}
		jlog.Logf("Checked outputs: %d, missing: %d\n", checkLockUtxos.CheckedOutputs, len(checkLockUtxos.MissingUtxos))
		for _, missingUtxo := range checkLockUtxos.MissingUtxos {
			jlog.Logf("unspent output without lock utxo: %s:%d\n",
				hs.GetTxString(missingUtxo.TxHash), missingUtxo.Index)
		}
	},
}
