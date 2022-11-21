package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/node/act/double_spend"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/config"
	"github.com/memocash/index/ref/dbi"
	"time"
)

type DoubleSpend struct {
	Verbose bool
}

func (s *DoubleSpend) SaveTxs(b *dbi.Block) error {
	if b.IsNil() {
		return jerr.Newf("error nil block")
	}
	block := b.ToWireBlock()
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
	var doubleSpendSeens []*item.DoubleSpendSeen
	for _, msgTx := range block.Transactions {
		txHash := msgTx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if s.Verbose {
			jlog.Logf("Double spend tx: %s\n", txHash.String())
		}
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
					doubleSpendSeens = append(doubleSpendSeens, &item.DoubleSpendSeen{
						TxHash:    prevOutHash,
						Index:     prevOutIndex,
						Timestamp: time.Now(),
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
	var doubleSpendOuts = make([]memo.Out, len(doubleSpendOutputs))
	for i := range doubleSpendOutputs {
		doubleSpendOuts[i] = memo.Out{
			TxHash: doubleSpendOutputs[i].TxHash,
			Index:  doubleSpendOutputs[i].Index,
		}
	}
	existingDoubleSpendOutputs, err := item.GetDoubleSpendsByOuts(doubleSpendOuts)
	if err != nil {
		return jerr.Get("error getting existing double spend outputs for double spends", err)
	}
	var numDoubleSpendInputs = len(doubleSpendInputs)
	if numDoubleSpendInputs > 0 {
		var objects = make([]db.Object, numDoubleSpendInputs+len(doubleSpendOutputs))
		for i := range doubleSpendInputs {
			objects[i] = doubleSpendInputs[i]
		}
		for i := range doubleSpendOutputs {
			objects[numDoubleSpendInputs+i] = doubleSpendOutputs[i]
		}
	DoubleSeenLoop:
		for _, doubleSpendSeen := range doubleSpendSeens {
			for _, existingDoubleSpendOutput := range existingDoubleSpendOutputs {
				if bytes.Equal(existingDoubleSpendOutput.TxHash, doubleSpendSeen.TxHash) &&
					existingDoubleSpendOutput.Index == doubleSpendSeen.Index {
					continue DoubleSeenLoop
				}
			}
			objects = append(objects, doubleSpendSeen)
		}
		if err = db.Save(objects); err != nil {
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
	if err := double_spend.AttachAllToDoubleSpendChecks(doubleSpendChecks); err != nil {
		return jerr.Get("error attaching all to double spend checks", err)
	}
	recentHeightBlock, err := item.GetRecentHeightBlock()
	if err != nil {
		return jerr.Get("error getting recent height block", err)
	}
	var recentHeight int64
	if recentHeightBlock != nil {
		recentHeight = recentHeightBlock.Height
	}
	blocksToConfirm := int64(config.GetBlocksToConfirm())
	var lostTxsToRemove []*item.TxLost
	var suspectTxsToRemove [][]byte
	var newItems []db.Object
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
			jlog.Logf("Checking double spend: %s:%d (out: %s:%d, win: %t)\n",
				hs.GetTxString(doubleSpendCheck.ParentTxHash), doubleSpendCheck.ParentTxIndex,
				hs.GetTxString(checkSpend.TxHash), checkSpend.Index, isWinner)
			var allTxHashes [][]byte
			var newTxHashes = [][]byte{checkSpend.TxHash}
			for len(newTxHashes) > 0 {
				txBlocks, err := item.GetTxBlocks(newTxHashes)
				if err != nil {
					return jerr.Get("error getting tx blocks for double spend check", err)
				}
				txLosts, err := item.GetTxLosts(newTxHashes)
				if err != nil {
					return jerr.Get("error getting tx losts for double spend check", err)
				}
				txSuspects, err := item.GetTxSuspects(newTxHashes)
				if err != nil {
					return jerr.Get("error getting tx suspects for double spend check", err)
				}
				txOutputs, err := item.GetTxOutputsByHashes(newTxHashes)
				if err != nil {
					return jerr.Get("error getting tx outputs for double spend check", err)
				}
				var outs = make([]memo.Out, len(txOutputs))
				for i := range txOutputs {
					outs[i] = memo.Out{
						TxHash:   txOutputs[i].TxHash,
						Index:    txOutputs[i].Index,
						Value:    txOutputs[i].Value,
						LockHash: txOutputs[i].LockHash,
					}
				}
				if isWinner {
					lockUtxoLosts, err := item.GetLockUtxoLosts(outs)
					if err != nil {
						return jerr.Get("error getting lock utxo losts for double spend check", err)
					}
					var lostOuts = make([]memo.Out, len(lockUtxoLosts))
					for i := range lockUtxoLosts {
						lostOuts[i] = memo.Out{
							TxHash: lockUtxoLosts[i].Hash,
							Index:  lockUtxoLosts[i].Index,
						}
					}
					outputInputs, err := item.GetOutputInputs(outs)
					if err != nil {
						return jerr.Get("error getting output inputs for lock utxo losts", err)
					}
				LockUtxoLostLoop:
					for _, lockUtxoLost := range lockUtxoLosts {
						for _, outputInput := range outputInputs {
							if bytes.Equal(outputInput.PrevHash, lockUtxoLost.Hash) && outputInput.PrevIndex == lockUtxoLost.Index {
								// Don't save LockUTXO if spent since lost saved
								continue LockUtxoLostLoop
							}
						}
						newItems = append(newItems, &item.LockUtxo{
							LockHash: lockUtxoLost.LockHash,
							Hash:     lockUtxoLost.Hash,
							Index:    lockUtxoLost.Index,
							Value:    lockUtxoLost.Value,
						})
					}
					if err := item.RemoveLockUtxoLosts(lockUtxoLosts); err != nil {
						return jerr.Get("error removing lock utxo losts for double spend check", err)
					}
				} else {
					lockUtxos, err := item.GetLockUtxosByOuts(outs)
					if err != nil {
						return jerr.Get("error getting lock utxos for double spend check", err)
					}
					for _, lockUtxo := range lockUtxos {
						newItems = append(newItems, &item.LockUtxoLost{
							LockHash: lockUtxo.LockHash,
							Hash:     lockUtxo.Hash,
							Index:    lockUtxo.Index,
							Value:    lockUtxo.Value,
						})
					}
					if err := item.RemoveLockUtxos(lockUtxos); err != nil {
						return jerr.Get("error removing lock utxos for double spend check", err)
					}
				}
				var blockHashes [][]byte
				for _, txBlock := range txBlocks {
					blockHashes = append(blockHashes, txBlock.BlockHash)
				}
				blockHeights, err := item.GetBlockHeights(blockHashes)
				if err != nil {
					return jerr.Get("error getting block heights for double spend check", err)
				}
				var needsChildTxHashes [][]byte
				for _, txHash := range newTxHashes {
					var isConfirmed, hasSuspect bool
					var foundTxLost *item.TxLost
					for _, txBlock := range txBlocks {
						if bytes.Equal(txBlock.TxHash, txHash) {
							for _, blockHeight := range blockHeights {
								if bytes.Equal(blockHeight.BlockHash, txBlock.BlockHash) {
									confirmations := recentHeight - blockHeight.Height
									isConfirmed = confirmations >= blocksToConfirm
									break
								}
							}
							break
						}
					}
					for _, txLost := range txLosts {
						if bytes.Equal(txLost.TxHash, txHash) {
							foundTxLost = txLost
							break
						}
					}
					for _, txSuspect := range txSuspects {
						if bytes.Equal(txSuspect.TxHash, txHash) {
							hasSuspect = true
							break
						}
					}
					if isWinner {
						if foundTxLost != nil {
							jlog.Logf("Removing TxLost: %s (spend: %s, parent: %s)\n", hs.GetTxString(txHash),
								hs.GetTxString(checkSpend.TxHash), hs.GetTxString(doubleSpendCheck.ParentTxHash))
							lostTxsToRemove = append(lostTxsToRemove, foundTxLost)
						}
						if !isConfirmed {
							newItems = append(newItems, &item.TxSuspect{
								TxHash: txHash,
							})
						} else if hasSuspect {
							suspectTxsToRemove = append(suspectTxsToRemove, txHash)
						}
					} else {
						if foundTxLost == nil {
							jlog.Logf("Adding TxLost from double spend: %s (spend: %s, parent: %s)\n", hs.GetTxString(txHash),
								hs.GetTxString(checkSpend.TxHash), hs.GetTxString(doubleSpendCheck.ParentTxHash))
							newItems = append(newItems, &item.TxLost{
								TxHash:      txHash,
								DoubleSpend: checkSpend.TxHash,
							})
						}
						if hasSuspect {
							suspectTxsToRemove = append(suspectTxsToRemove, txHash)
						}
					}
					if !isConfirmed || foundTxLost != nil || hasSuspect || !isWinner {
						needsChildTxHashes = append(needsChildTxHashes, txHash)
					}
				}
				newTxHashes = nil
				outputInputs, err := item.GetOutputInputsForTxHashes(needsChildTxHashes)
				if err != nil {
					return jerr.Get("error getting output inputs for tx hash descendants", err)
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
			descendantLockHashes, err := GetTxLockHashes(allTxHashes)
			if err != nil {
				return jerr.Get("error getting tx lock hashes for descendant tx hashes double spend", err)
			}
			lockHashes = append(lockHashes, descendantLockHashes...)
		}
	}
	if err := db.Save(newItems); err != nil {
		return jerr.Get("error saving new item tx suspects and tx losts", err)
	}
	if err := item.RemoveTxLosts(lostTxsToRemove); err != nil {
		return jerr.Get("error removing tx losts for winner", err)
	}
	if err := item.RemoveTxSuspects(suspectTxsToRemove); err != nil {
		return jerr.Get("error removing tx suspects for double spends", err)
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
	var newItemObjects []db.Object
	var newTxLosts []item.TxLost
	for _, txLost := range parentTxLosts {
	LostTxLoop:
		for _, tx := range txs {
			for _, in := range tx.TxIn {
				if bytes.Equal(in.PreviousOutPoint.Hash.CloneBytes(), txLost.TxHash) {
					txHash := tx.TxHash()
					txHashBytes := txHash.CloneBytes()
					for _, newTxLost := range newTxLosts {
						if bytes.Equal(newTxLost.TxHash, txHashBytes) &&
							bytes.Equal(newTxLost.DoubleSpend, txLost.DoubleSpend) {
							continue LostTxLoop
						}
					}
					var parentTxHash []byte
					if len(txLost.DoubleSpend) > 0 {
						parentTxHash = txLost.DoubleSpend
					} else {
						parentTxHash = txLost.TxHash
					}
					jlog.Logf("Adding TxLost from Parent: %s (parent: %s %s)\n",
						txHash.String(), hs.GetTxString(txLost.TxHash), hs.GetTxString(txLost.DoubleSpend))
					newTxLosts = append(newTxLosts, item.TxLost{
						TxHash:      txHashBytes,
						DoubleSpend: parentTxHash,
					})
					continue LostTxLoop
				}
			}
		}
	}
	for i := range newTxLosts {
		newItemObjects = append(newItemObjects, &newTxLosts[i])
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
	var maxHeight int64
	var blocksToConfirm int64
	var parentBlockHeights []*item.BlockHeight
	if len(parentBlockHeightsToGet) > 0 {
		recentHeightBlock, err := item.GetRecentHeightBlock()
		if err != nil {
			return jerr.Get("error getting recent height block", err)
		}
		maxHeight = recentHeightBlock.Height
		if parentBlockHeights, err = item.GetBlockHeights(parentBlockHeightsToGet); err != nil {
			return jerr.Get("error getting block heights for double spends", err)
		}
		blocksToConfirm = int64(config.GetBlocksToConfirm())
	}
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
						if maxHeight-parentBlockHeight.Height >= blocksToConfirm {
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
	if err := db.Save(newItemObjects); err != nil {
		return jerr.Get("error saving new tx losts", err)
	}
	return nil
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
