package maint

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/spf13/cobra"
)

var populateDoubleSpendSeenCmd = &cobra.Command{
	Use: "populate-double-spend-seen",
	Run: func(cmd *cobra.Command, args []string) {
		var startTxHash []byte
		for {
			doubleSpendOutputs, err := item.GetDoubleSpendOutputs(startTxHash, client.DefaultLimit)
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
			for i := range txSeens {
				for _, newDoubleSpendSeen := range newDoubleSpendSeens {
					if bytes.Equal(newDoubleSpendSeen.TxHash, txSeens[i].TxHash) {
						newDoubleSpendSeen.Timestamp = txSeens[i].Timestamp
						objects[i] = newDoubleSpendSeen
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
			startTxHash = doubleSpendOutputs[len(doubleSpendOutputs)-1].TxHash
		}
	},
}
