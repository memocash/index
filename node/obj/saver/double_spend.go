package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/node/act/double_spend"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"github.com/memocash/server/ref/config"
)

type DoubleSpend struct {
	Verbose bool
}

func (s *DoubleSpend) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var inputOuts []memo.Out
	for _, msgTx := range block.Transactions {
		for _, in := range msgTx.TxIn {
			inputOuts = append(inputOuts, memo.Out{
				TxHash: in.PreviousOutPoint.Hash.CloneBytes(),
				Index:  in.PreviousOutPoint.Index,
			})
		}
	}
	var blockHashBytes []byte
	if !block.Header.Timestamp.IsZero() {
		blockHash := block.BlockHash()
		blockHashBytes = blockHash.CloneBytes()
	}
	existingOutputInputs, err := item.GetOutputInputs(inputOuts)
	if err != nil {
		return jerr.Get("error getting output inputs", err)
	}
	var doubleSpendChecks []*double_spend.DoubleSpendCheck
	var doubleSpendInputs []*item.DoubleSpendInput
	var doubleSpendOutputs []*item.DoubleSpendOutput
	for _, msgTx := range block.Transactions {
		txHash := msgTx.TxHash()
		txHashBytes := txHash.CloneBytes()
		for index, in := range msgTx.TxIn {
			prevOutHash := in.PreviousOutPoint.Hash.CloneBytes()
			prevOutIndex := in.PreviousOutPoint.Index
			for _, ex := range existingOutputInputs {
				if (bytes.Equal(prevOutHash, ex.PrevHash) && prevOutIndex == ex.PrevIndex) &&
					(!bytes.Equal(txHashBytes, ex.Hash) || uint32(index) != ex.Index) {
					// Double Spend!
					doubleSpendInputs = append(doubleSpendInputs, &item.DoubleSpendInput{
						TxHash: txHashBytes,
						Index:  uint32(index),
					}, &item.DoubleSpendInput{
						TxHash: ex.Hash,
						Index:  ex.Index,
					})
					doubleSpendOutputs = append(doubleSpendOutputs, &item.DoubleSpendOutput{
						TxHash: prevOutHash,
						Index:  prevOutIndex,
					})
					doubleSpendChecks = append(doubleSpendChecks, &double_spend.DoubleSpendCheck{
						ParentTxHash:  prevOutHash,
						ParentTxIndex: prevOutIndex,
						Spends: []*double_spend.DoubleSpendCheckSpend{{
							TxHash:    txHashBytes,
							Index:     uint32(index),
							BlockHash: blockHashBytes,
						}, {
							TxHash: ex.Hash,
							Index:  ex.Index,
						}},
					})
				}
			}
		}
	}
	var numDoubleSpendInputs = len(doubleSpendInputs)
	if numDoubleSpendInputs > 0 {
		var objects = make([]item.Object, numDoubleSpendInputs+len(doubleSpendOutputs))
		for i := range doubleSpendInputs {
			objects[i] = doubleSpendInputs[i]
		}
		for i := range doubleSpendOutputs {
			objects[numDoubleSpendInputs+i] = doubleSpendOutputs[i]
		}
		if err = item.Save(objects); err != nil {
			return jerr.Get("error saving double spend objects", err)
		}
	}
	if err := s.CheckLost(doubleSpendChecks); err != nil {
		return jerr.Get("error checking lost txs for double spend", err)
	}
	if err := s.AddLostAndSuspectByParents(block.Transactions); err != nil {
		return jerr.Get("error adding lost and suspect by parents double spends", err)
	}
	return nil
}

func (s *DoubleSpend) CheckLost(doubleSpendChecks []*double_spend.DoubleSpendCheck) error {
	var txHashes [][]byte
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, spend := range doubleSpendCheck.Spends {
			txHashes = append(txHashes, spend.TxHash)
		}
	}
	txHashes = jutil.RemoveDupesAndEmpties(txHashes)
	var itemTxSuspects = make([]item.Object, len(txHashes))
	for i := range txHashes {
		itemTxSuspects[i] = &item.TxSuspect{
			TxHash: txHashes[i],
		}
	}
	if err := item.Save(itemTxSuspects); err != nil {
		return jerr.Get("error saving item tx suspects", err)
	}
	if err := double_spend.AttachAllToDoubleSpendChecks(doubleSpendChecks); err != nil {
		return jerr.Get("error attaching all to double spend checks", err)
	}
	var lostTxsToRemove [][]byte
	var newTxLosts []item.Object
	var lockHashes [][]byte
	for _, doubleSpendCheck := range doubleSpendChecks {
		lockHashes = append(lockHashes, doubleSpendCheck.LockHash)
		for _, checkSpend := range doubleSpendCheck.Spends {
			lockHashes = append(lockHashes, checkSpend.LockHashes...)
			isWinner, err := doubleSpendCheck.IsWinnerSpend(checkSpend)
			if err != nil {
				return jerr.Getf(err, "error checking if double spend check is winner (%s:%d)",
					hs.GetTxString(doubleSpendCheck.ParentTxHash), doubleSpendCheck.ParentTxIndex)
			}
			descendantTxHashes, err := GetTxDescendants(checkSpend.TxHash)
			if err != nil {
				return jerr.Get("error getting descendant tx hashes for double spend", err)
			}
			if isWinner {
				lostTxsToRemove = append(lostTxsToRemove, checkSpend.TxHash)
				lostTxsToRemove = append(lostTxsToRemove, descendantTxHashes...)
			} else {
				newTxLosts = append(newTxLosts, &item.TxLost{
					TxHash: checkSpend.TxHash,
				})
				for _, descendantTxHash := range descendantTxHashes {
					newTxLosts = append(newTxLosts, &item.TxLost{
						TxHash: descendantTxHash,
					})
				}
			}
			descendantLockHashes, err := GetTxLockHashes(descendantTxHashes)
			if err != nil {
				return jerr.Get("error getting tx lock hashes for descendant tx hashes double spend", err)
			}
			lockHashes = append(lockHashes, descendantLockHashes...)
		}
	}
	if err := item.RemoveTxLosts(lostTxsToRemove); err != nil {
		return jerr.Get("error removing tx losts for winner", err)
	}
	if err := item.Save(newTxLosts); err != nil {
		return jerr.Get("error saving new tx losts", err)
	}
	if err := item.RemoveLockBalances(lockHashes); err != nil {
		return jerr.Get("error removing lock balances", err)
	}
	return nil
}

func (s *DoubleSpend) AddLostAndSuspectByParents(txs []*wire.MsgTx) error {
	var parentTxHashes [][]byte
	for _, tx := range txs {
		for _, in := range tx.TxIn {
			parentTxHashes = append(parentTxHashes, in.PreviousOutPoint.Hash.CloneBytes())
		}
	}
	parentTxLosts, err := item.GetTxLosts(parentTxHashes)
	if err != nil {
		return jerr.Get("error getting tx losts for double spend check txs", err)
	}
	var newItemObjects []item.Object
	for _, txLost := range parentTxLosts {
		for _, tx := range txs {
			for _, in := range tx.TxIn {
				if bytes.Equal(in.PreviousOutPoint.Hash.CloneBytes(), txLost.TxHash) {
					txHash := tx.TxHash()
					newItemObjects = append(newItemObjects, &item.TxLost{
						TxHash: txHash.CloneBytes(),
					})
				}
			}
		}
	}
	parentTxSuspects, err := item.GetTxSuspects(parentTxHashes)
	if err != nil {
		return jerr.Get("error getting tx suspects for double spend check txs", err)
	}
	var parentTxSuspectHashes = make([][]byte, len(parentTxSuspects))
	for i := range parentTxSuspects {
		parentTxSuspectHashes[i] = parentTxSuspects[i].TxHash
	}
	parentTxSuspectBlocks, err := item.GetTxBlocks(parentTxSuspectHashes)
	if err != nil {
		return jerr.Get("error getting tx blocks for suspect tx hashes", err)
	}
	var parentBlockHeightsToGet [][]byte
	for i := 0; i < len(parentTxSuspects); i++ {
		for _, parentTxSuspectBlock := range parentTxSuspectBlocks {
			if bytes.Equal(parentTxSuspectBlock.TxHash, parentTxSuspects[i].TxHash) {
				parentBlockHeightsToGet = append(parentBlockHeightsToGet, parentTxSuspectBlock.BlockHash)
				break
			}
		}
	}
	if len(parentBlockHeightsToGet) > 0 {
		recentHeightBlock, err := item.GetRecentHeightBlock()
		if err != nil {
			return jerr.Get("error getting recent height block", err)
		}
		maxHeight := recentHeightBlock.Height
		parentBlockHeights, err := item.GetBlockHeights(parentBlockHeightsToGet)
		if err != nil {
			return jerr.Get("error getting block heights for double spends", err)
		}
		blocksToConfirm := config.GetBlocksToConfirm()
	TxLoop:
		for _, tx := range txs {
		InputLoop:
			for _, in := range tx.TxIn {
				var parentTxSuspectFound *item.TxSuspect
				for _, parentTxSuspect := range parentTxSuspects {
					if bytes.Equal(parentTxSuspect.TxHash, in.PreviousOutPoint.Hash.CloneBytes()) {
						parentTxSuspectFound = parentTxSuspect
						break
					}
				}
				if parentTxSuspectFound == nil {
					continue
				}
				var parentTxSuspectBlockFound *item.TxBlock
				for _, parentTxSuspectBlock := range parentTxSuspectBlocks {
					if bytes.Equal(parentTxSuspectBlock.TxHash, parentTxSuspectFound.TxHash) {
						parentTxSuspectBlockFound = parentTxSuspectBlock
						break
					}
				}
				if parentTxSuspectBlockFound != nil {
					for _, parentBlockHeight := range parentBlockHeights {
						if bytes.Equal(parentBlockHeight.BlockHash, parentTxSuspectBlockFound.BlockHash) {
							if maxHeight-parentBlockHeight.Height >= int64(blocksToConfirm) {
								// Don't add suspect if parent old
								continue InputLoop
							}
							break
						}
					}
				}
				txHash := tx.TxHash()
				newItemObjects = append(newItemObjects, &item.TxSuspect{
					TxHash: txHash.CloneBytes(),
				})
				continue TxLoop
			}
		}
	}
	if err := item.Save(newItemObjects); err != nil {
		return jerr.Get("error saving new tx losts", err)
	}
	return nil
}

func GetTxDescendants(txHash []byte) ([][]byte, error) {
	var allTxHashes [][]byte
	var newTxHashes = [][]byte{txHash}
	for len(newTxHashes) > 0 {
		var txHashesToCheck = newTxHashes
		newTxHashes = [][]byte{}
		outputInputs, err := item.GetOutputInputsForTxHashes(txHashesToCheck)
		if err != nil {
			return nil, jerr.Get("error getting output inputs for tx hash descendants", err)
		}
	Loop:
		for _, outputInput := range outputInputs {
			for _, allTxHash := range allTxHashes {
				if bytes.Equal(allTxHash, outputInput.Hash) {
					continue Loop
				}
			}
			allTxHashes = append(allTxHashes, outputInput.Hash)
			newTxHashes = append(newTxHashes, outputInput.Hash)
		}
	}
	return allTxHashes, nil
}

func GetTxLockHashes(txHashes [][]byte) ([][]byte, error) {
	txInputs, err := item.GetTxInputsByHashes(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx inputs for lock hashes", err)
	}
	var outs = make([]memo.Out, len(txInputs))
	for i := range txInputs {
		outs[i] = memo.Out{
			TxHash: txInputs[i].PrevHash,
			Index:  txInputs[i].PrevIndex,
		}
	}
	txOutputs, err := item.GetTxOutputsByHashes(txHashes)
	if err != nil {
		return nil, jerr.Get("error getting tx outputs for lock hashes txs", err)
	}
	txOutputsFromInputs, err := item.GetTxOutputs(outs)
	if err != nil {
		return nil, jerr.Get("error getting tx outputs for lock hashes inputs", err)
	}
	var lenTxOutputs = len(txOutputs)
	var lockHashes = make([][]byte, lenTxOutputs+len(txOutputsFromInputs))
	for i := range txOutputs {
		lockHashes[i] = txOutputs[i].LockHash
	}
	for i := range txOutputsFromInputs {
		lockHashes[lenTxOutputs+i] = txOutputsFromInputs[i].LockHash
	}
	return lockHashes, nil
}

func NewDoubleSpend(verbose bool) *DoubleSpend {
	return &DoubleSpend{
		Verbose: verbose,
	}
}
