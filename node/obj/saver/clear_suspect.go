package saver

import (
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/config"
)

type ClearSuspect struct {
	Verbose bool
}

func (s *ClearSuspect) SaveBlock(block wire.BlockHeader) error {
	saveBlockHash := block.BlockHash()
	saveBlockHeight, err := item.GetBlockHeight(saveBlockHash.CloneBytes())
	if err != nil && !client.IsEntryNotFoundError(err) {
		return jerr.Get("error getting block height for clear suspect", err)
	}
	if saveBlockHeight == nil {
		return nil
	}
	blocksToConfirm := config.GetBlocksToConfirm()
	if saveBlockHeight == nil || saveBlockHeight.Height <= int64(blocksToConfirm) {
		return nil
	}
	confirmedHeightBlocks, err := item.GetHeightBlock(saveBlockHeight.Height - int64(blocksToConfirm))
	if err != nil {
		return jerr.Get("error getting height block for confirm to clear suspect", err)
	}
	if len(confirmedHeightBlocks) != 1 {
		return jerr.Newf("error unexpected number of height blocks returned for clear suspect: %d",
			len(confirmedHeightBlocks))
	}
	const limit = client.DefaultLimit
	var blockHash = confirmedHeightBlocks[0].BlockHash
	var startUid []byte
	for {
		blockTxes, err := item.GetBlockTxes(item.BlockTxesRequest{
			BlockHash: blockHash,
			StartUid:  startUid,
			Limit:     limit,
		})
		if err != nil {
			return jerr.Get("error getting block txs for clear suspect", err)
		}
		var txHashes = make([][]byte, len(blockTxes))
		for i := range blockTxes {
			txHashes[i] = blockTxes[i].TxHash
		}
		doubleSpendInputs, err := item.GetDoubleSpendInputsByTxHashes(txHashes)
		if err != nil {
			return jerr.Get("error getting double spend inputs by tx hashes", err)
		}
		var inputTxsToClear = make([][]byte, len(doubleSpendInputs))
		for i := range doubleSpendInputs {
			inputTxsToClear[i] = doubleSpendInputs[i].TxHash
		}
		if err := s.ClearSuspectAndDescendants(inputTxsToClear, true); err != nil {
			return jerr.Get("error clearing suspect and descendants", err)
		}
		if len(blockTxes) < limit {
			break
		}
		startUid = item.GetBlockTxUid(blockHash, blockTxes[len(blockTxes)-1].TxHash)
	}
	return nil
}

func (s *ClearSuspect) ClearSuspectAndDescendants(txHashes [][]byte, checkHasSuspect bool) error {
	for i := 0; len(txHashes) > 0; i++ {
		var processTxHashes = txHashes
		txHashes = nil
		var removeSuspectTxHashes [][]byte
		if checkHasSuspect {
			txSuspects, err := item.GetTxSuspects(processTxHashes)
			if err != nil {
				return jerr.Getf(err, "error getting tx suspects for process clear suspect txs (loop: %d)", i)
			}
			removeSuspectTxHashes = make([][]byte, len(txSuspects))
			for i := range txSuspects {
				removeSuspectTxHashes[i] = txSuspects[i].TxHash
			}
		} else {
			removeSuspectTxHashes = processTxHashes
		}
		if err := item.RemoveTxSuspects(removeSuspectTxHashes); err != nil {
			return jerr.Get("error removing suspect txs", err)
		}
		outputInputs, err := item.GetOutputInputsForTxHashes(removeSuspectTxHashes)
		if err != nil {
			return jerr.Get("error getting output inputs for clear suspect tx hash descendants", err)
		}
		for _, outputInput := range outputInputs {
			txHashes = append(txHashes, outputInput.Hash)
		}
	}
	return nil
}

func (s *ClearSuspect) GetBlock(int64) ([]byte, error) {
	return nil, nil
}

func NewClearSuspect(verbose bool) *ClearSuspect {
	return &ClearSuspect{
		Verbose: verbose,
	}
}
