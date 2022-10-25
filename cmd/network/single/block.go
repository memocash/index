package single

import (
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/act/block_tx"
	"github.com/memocash/index/node/obj/saver"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/dbi"
	"github.com/spf13/cobra"
)

var doubleSpendBlockCmd = &cobra.Command{
	Use:   "double-spend-block",
	Short: "double-spend-block BLOCK_HASH",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			jerr.New("fatal error must specify block hash").Fatal()
		}
		blockHash, err := chainhash.NewHashFromStr(args[0])
		if err != nil {
			jerr.Get("fatal error parsing block hash", err).Fatal()
		}
		blockHashBytes := blockHash.CloneBytes()
		block, err := item.GetBlock(blockHashBytes)
		if err != nil {
			jerr.Get("fatal error getting block", err).Fatal()
		}
		blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
		if err != nil {
			jerr.Get("fatal error getting block header from raw", err).Fatal()
		}
		doubleSpendSaver := saver.NewDoubleSpend(false)
		if err := block_tx.NewLoopRaw(func(blockTxesRaw []*item.BlockTxRaw) error {
			var msgTxs = make([]*wire.MsgTx, len(blockTxesRaw))
			for i := range blockTxesRaw {
				msgTxs[i], err = memo.GetMsgFromRaw(blockTxesRaw[i].Raw)
				if err != nil {
					return jerr.Get("error getting tx from raw block tx", err)
				}
			}
			err = doubleSpendSaver.SaveTxs(dbi.GetBlock(memo.GetBlockFromTxs(msgTxs, blockHeader)))
			return nil
		}).Process(blockHashBytes); err != nil {
			jerr.Get("fatal error processing block txs for double spend", err).Fatal()
		}
	},
}

var lockHeightBlockCmd = &cobra.Command{
	Use:   "lock-height-block",
	Short: "lock-height-block BLOCK_HASH TX_HASH",
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			jerr.New("fatal error must specify block hash").Fatal()
		}
		blockHash, err := chainhash.NewHashFromStr(args[0])
		if err != nil {
			jerr.Get("fatal error parsing block hash", err).Fatal()
		}
		var checkTxHashBytes []byte
		if len(args) >= 2 {
			checkTxHash, err := chainhash.NewHashFromStr(args[1])
			if err != nil {
				jerr.Get("fatal error parsing tx hash", err)
			}
			checkTxHashBytes = checkTxHash.CloneBytes()
		}
		blockHashBytes := blockHash.CloneBytes()
		block, err := item.GetBlock(blockHashBytes)
		if err != nil {
			jerr.Get("fatal error getting block", err).Fatal()
		}
		blockHeader, err := memo.GetBlockHeaderFromRaw(block.Raw)
		if err != nil {
			jerr.Get("fatal error getting block header from raw", err).Fatal()
		}
		lockHeightSaver := saver.NewLockHeight(true)
		lockHeightSaver.CheckTxHash = checkTxHashBytes
		if err := block_tx.NewLoopRaw(func(blockTxesRaw []*item.BlockTxRaw) error {
			var msgTxs = make([]*wire.MsgTx, len(blockTxesRaw))
			for i := range blockTxesRaw {
				msgTxs[i], err = memo.GetMsgFromRaw(blockTxesRaw[i].Raw)
				if err != nil {
					return jerr.Get("error getting tx from raw block tx", err)
				}
			}
			err = lockHeightSaver.SaveTxs(dbi.GetBlock(memo.GetBlockFromTxs(msgTxs, blockHeader)))
			return nil
		}).Process(blockHashBytes); err != nil {
			jerr.Get("fatal error processing block txs for lock heights", err).Fatal()
		}
	},
}
