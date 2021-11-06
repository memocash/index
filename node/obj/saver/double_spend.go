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
	return nil
}

func (s *DoubleSpend) CheckLost(doubleSpendChecks []*double_spend.DoubleSpendCheck) error {
	var txHashes [][]byte
	for _, doubleSpendCheck := range doubleSpendChecks {
		txHashes = append(txHashes, doubleSpendCheck.ParentTxHash)
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
			if isWinner {
				lostTxsToRemove = append(lostTxsToRemove, checkSpend.TxHash)
				// TODO: Recursively remove existing tx_losts for children
			} else {
				newTxLosts = append(newTxLosts, &item.TxLost{
					TxHash: checkSpend.TxHash,
				})
				// TODO: Recursively add tx losts for children
			}
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

func NewDoubleSpend(verbose bool) *DoubleSpend {
	return &DoubleSpend{
		Verbose: verbose,
	}
}
