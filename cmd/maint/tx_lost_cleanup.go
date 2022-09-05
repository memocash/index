package maint

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/config"
	"github.com/spf13/cobra"
)

var txLostCleanupCmd = &cobra.Command{
	Use: "tx-lost",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool(FlagVerbose)
		var totalTxLosts int
		var txLostsRemoved int
		currentHeightBlock, err := item.GetRecentHeightBlock()
		if err != nil {
			jerr.Get("fatal error getting recent height block", err).Fatal()
		}
		confirmsRequired := config.GetBlocksToConfirm()
		for _, shard := range config.GetQueueShards() {
			var lastUid []byte
			for {
				txLosts, err := item.GetAllTxLosts(shard.Shard, lastUid)
				if err != nil {
					jerr.Get("fatal error getting all tx losts for cleanup", err).Fatal()
				}
				if len(lastUid) > 0 && len(txLosts) > 0 && bytes.Equal(txLosts[0].GetUid(), lastUid) {
					txLosts = txLosts[1:]
				}
				var txHashes = make([][]byte, len(txLosts))
				for i := range txLosts {
					if len(txLosts[i].DoubleSpend) > 0 {
						txHashes[i] = txLosts[i].DoubleSpend
					} else {
						txHashes[i] = txLosts[i].TxHash
					}
				}
				txHashes = jutil.RemoveDupesAndEmpties(txHashes)
				txBlocks, err := item.GetTxBlocks(txHashes)
				if err != nil {
					jerr.Get("fatal error getting tx blocks for tx losts maint", err).Fatal()
				}
				var blockHashes = make([][]byte, len(txBlocks))
				for i := range txBlocks {
					blockHashes[i] = txBlocks[i].BlockHash
				}
				blockHeights, err := item.GetBlockHeights(blockHashes)
				if err != nil {
					jerr.Get("fatal error getting block heights", err).Fatal()
				}
				for i := 0; i < len(txBlocks); i++ {
					var blockHeightFound *item.BlockHeight
					for _, blockHeight := range blockHeights {
						if bytes.Equal(blockHeight.BlockHash, txBlocks[i].BlockHash) {
							blockHeightFound = blockHeight
							break
						}
					}
					if blockHeightFound == nil || currentHeightBlock.Height-blockHeightFound.Height < int64(confirmsRequired) {
						// Block found but not considered confirmed yet
						txBlocks = append(txBlocks[:i], txBlocks[i+1:]...)
						i--
					}
				}
				var txLostsToRemove []*item.TxLost
				for _, txLost := range txLosts {
					var txLostTxHash []byte
					if len(txLost.DoubleSpend) > 0 {
						txLostTxHash = txLost.DoubleSpend
					} else {
						txLostTxHash = txLost.TxHash
					}
					for _, txBlock := range txBlocks {
						if bytes.Equal(txLostTxHash, txBlock.TxHash) {
							txLostsToRemove = append(txLostsToRemove, txLost)
							if verbose {
								jlog.Logf("Removing TxLost: %s (ds: %s)\n",
									hs.GetTxString(txLost.TxHash), hs.GetTxString(txLost.DoubleSpend))
							}
							break
						}
					}
				}
				var txLostsToRemoveHashes = make([][]byte, len(txLostsToRemove))
				for i := range txLostsToRemove {
					txLostsToRemoveHashes[i] = txLostsToRemove[i].TxHash
				}
				lockHashBalancesToRemove, err := saver.GetTxLockHashes(txLostsToRemoveHashes)
				if err != nil {
					jerr.Get("fatal error getting tx lock hashes", err).Fatal()
				}
				if err := item.RemoveLockBalances(lockHashBalancesToRemove); err != nil {
					jerr.Get("fatal error removing lock balances", err)
				}
				if err := item.RemoveTxLosts(txLostsToRemove); err != nil {
					jerr.Get("fatal error removing tx losts for maint", err).Fatal()
				}
				totalTxLosts += len(txLosts)
				txLostsRemoved += len(txLostsToRemove)
				if len(txLosts) < client.DefaultLimit-1 {
					break
				}
				jlog.Logf("len(txLosts): %d, len(txLostsToRemove): %d, len(lockHashBalancesToRemove): %d, Last: %s\n",
					len(txLosts), len(txLostsToRemove), len(lockHashBalancesToRemove),
					hs.GetTxString(txLosts[len(txLosts)-1].TxHash))
				lastUid = txLosts[len(txLosts)-1].GetUid()
			}
		}
		jlog.Logf("TotalTxLosts: %d, TxLostsRemoved: %d\n", totalTxLosts, txLostsRemoved)
	},
}
