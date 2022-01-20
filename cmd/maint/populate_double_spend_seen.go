package maint

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/spf13/cobra"
)

var populateDoubleSpendSeenCmd = &cobra.Command{
	Use: "populate-double-spend-seen",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		var startDoubleSpendOutput *item.DoubleSpendOutput
		for {
			doubleSpendOutputs, err := item.GetDoubleSpendOutputs(startDoubleSpendOutput, client.DefaultLimit)
			if err != nil {
				jerr.Get("fatal error getting double spend outputs for seen population", err).Fatal()
			}
			var doubleSpendOutputTxHashes = make([][]byte, len(doubleSpendOutputs))
			for i := range doubleSpendOutputs {
				doubleSpendOutputTxHashes[i] = doubleSpendOutputs[i].TxHash
			}
			existingDoubleSpendSeens, err := item.GetDoubleSpendSeensByTxHashesScanAll(doubleSpendOutputTxHashes)
			if err != nil {
				jerr.Get("fatal error getting existing double spend seens for populate script", err).Fatal()
			}
			var newDoubleSpendSeens []*item.DoubleSpendSeen
			var txSeenTxHashes [][]byte
		DoubleSpendOutputsLoop:
			for _, doubleSpendOutput := range doubleSpendOutputs {
				for _, existingDoubleSpendSeen := range existingDoubleSpendSeens {
					if bytes.Equal(existingDoubleSpendSeen.TxHash, doubleSpendOutput.TxHash) {
						if verbose {
							jlog.Logf("Existing Double Spend Seen: %s:%-3d - %s (shard: %d)\n",
								hs.GetTxString(existingDoubleSpendSeen.TxHash), existingDoubleSpendSeen.Index,
								existingDoubleSpendSeen.Timestamp, item.GetShard(existingDoubleSpendSeen.GetShard()))
						}
						continue DoubleSpendOutputsLoop
					}
				}
				newDoubleSpendSeens = append(newDoubleSpendSeens, &item.DoubleSpendSeen{
					TxHash: doubleSpendOutput.TxHash,
					Index:  doubleSpendOutput.Index,
				})
				txSeenTxHashes = append(txSeenTxHashes, doubleSpendOutput.TxHash)
			}
			txSeens, err := item.GetTxSeens(txSeenTxHashes)
			if err != nil {
				jerr.Get("fatal error getting tx seens for populate double spend seens", err).Fatal()
			}
			var objects []item.Object
			for _, newDoubleSpendSeen := range newDoubleSpendSeens {
				for i := range txSeens {
					if bytes.Equal(txSeens[i].TxHash, newDoubleSpendSeen.TxHash) {
						newDoubleSpendSeen.Timestamp = txSeens[i].Timestamp
						if verbose {
							jlog.Logf("New Double Spend Seen: %s:%-3d - %s (shard: %d)\n",
								hs.GetTxString(newDoubleSpendSeen.TxHash), newDoubleSpendSeen.Index,
								newDoubleSpendSeen.Timestamp, item.GetShard(newDoubleSpendSeen.GetShard()))
						}
						objects = append(objects, newDoubleSpendSeen)
						break
					}
				}
			}
			jlog.Logf("Saving %d new double spend seens\n", len(objects))
			if err := item.Save(objects); err != nil {
				jerr.Get("fatal error saving new double spend seens for population", err).Fatal()
			}
			if len(doubleSpendOutputs) < client.DefaultLimit {
				break
			}
			startDoubleSpendOutput = doubleSpendOutputs[len(doubleSpendOutputs)-1]
		}
	},
}
