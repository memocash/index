package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/item"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/hs"
	"time"
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
	existingOutputInputs, err := item.GetOutputInputs(inputOuts)
	if err != nil {
		return jerr.Get("error getting output inputs", err)
	}
	var txSuspects []*memo.InOut
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
					txSuspects = append(txSuspects, &memo.InOut{
						Hash:      txHashBytes,
						Index:     uint32(index),
						PrevHash:  prevOutHash,
						PrevIndex: prevOutIndex,
					}, &memo.InOut{
						Hash:      ex.Hash,
						Index:     ex.Index,
						PrevHash:  prevOutHash,
						PrevIndex: prevOutIndex,
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
	if err := s.CheckLost(txSuspects); err != nil {
		return jerr.Get("error checking lost txs for double spend", err)
	}
	return nil
}

func (s *DoubleSpend) CheckLost(txSuspects []*memo.InOut) error {
	var txHashes [][]byte
	var doubleSpendChecks []*DoubleSpendCheck
SuspectLoop:
	for _, txSuspect := range txSuspects {
		txHashes = append(txHashes, txSuspect.Hash, txSuspect.PrevHash)
		var spend = &DoubleSpendCheckSpend{
			TxHash: txSuspect.Hash,
			Index:  txSuspect.Index,
		}
		for _, doubleSpendCheck := range doubleSpendChecks {
			if bytes.Equal(doubleSpendCheck.ParentTxHash, txSuspect.PrevHash) &&
				doubleSpendCheck.ParentTxIndex == txSuspect.PrevIndex {
				doubleSpendCheck.Spends = append(doubleSpendCheck.Spends, spend)
				continue SuspectLoop
			}
		}
		doubleSpendChecks = append(doubleSpendChecks, &DoubleSpendCheck{
			ParentTxHash:  txSuspect.PrevHash,
			ParentTxIndex: txSuspect.PrevIndex,
			Spends:        []*DoubleSpendCheckSpend{spend},
		})
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
	if err := AttachFirstSeenAndConfirmsToDoubleSpendCheckSpends(doubleSpendChecks); err != nil {
		return jerr.Get("error attaching first seen and confirms to double spend check spends", err)
	}
	var invalidTxsToRemove [][]byte
	var newTxInvalids []item.Object
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, checkSpend := range doubleSpendCheck.Spends {
			isWinner, err := doubleSpendCheck.IsWinnerSpend(checkSpend)
			if err != nil {
				return jerr.Getf(err, "error checking if double spend check is winner (%s:%d)",
					hs.GetTxString(doubleSpendCheck.ParentTxHash), doubleSpendCheck.ParentTxIndex)
			}
			if isWinner {
				invalidTxsToRemove = append(invalidTxsToRemove, checkSpend.TxHash)
				// TODO: Recursively remove existing tx_invalids for children
			} else {
				newTxInvalids = append(newTxInvalids, &item.TxInvalid{
					TxHash: checkSpend.TxHash,
				})
				// TODO: Recursively add tx invalids for children
			}
		}
	}
	if err := item.RemoveTxInvalids(invalidTxsToRemove); err != nil {
		return jerr.Get("error removing tx invalids for winner", err)
	}
	if err := item.Save(newTxInvalids); err != nil {
		return jerr.Get("error saving new tx invalids", err)
	}
	return nil
}

func NewDoubleSpend(verbose bool) *DoubleSpend {
	return &DoubleSpend{
		Verbose: verbose,
	}
}

func AttachFirstSeenAndConfirmsToDoubleSpendCheckSpends(doubleSpendChecks []*DoubleSpendCheck) error {
	var txHashes [][]byte
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, spend := range doubleSpendCheck.Spends {
			txHashes = append(txHashes, spend.TxHash)
		}
	}
	txSeens, err := item.GetTxSeens(txHashes)
	if err != nil {
		return jerr.Get("error getting tx seens for double spend check spends", err)
	}
	txBlocks, err := item.GetTxBlocks(txHashes)
	if err != nil {
		return jerr.Get("error getting tx blocks for double spend check spends", err)
	}
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, spend := range doubleSpendCheck.Spends {
			for _, txSeen := range txSeens {
				if bytes.Equal(txSeen.TxHash, spend.TxHash) {
					spend.FirstSeen = txSeen.Timestamp
					break
				}
			}
			for _, txBlock := range txBlocks {
				if bytes.Equal(txBlock.TxHash, spend.TxHash) {
					spend.BlockHash = txBlock.BlockHash
					break
				}
			}
		}
	}
	return nil
}

type DoubleSpendCheck struct {
	ParentTxHash  []byte
	ParentTxIndex uint32
	Spends        []*DoubleSpendCheckSpend
}

func (c DoubleSpendCheck) IsWinnerSpend(spendCheck *DoubleSpendCheckSpend) (bool, error) {
	for _, spend := range c.Spends {
		if bytes.Equal(spend.TxHash, spendCheck.TxHash) {
			continue
		}
		if len(spend.BlockHash) > 0 && len(spendCheck.BlockHash) == 0 {
			return false, nil
		}
		if len(spend.BlockHash) == 0 && len(spendCheck.BlockHash) > 0 {
			return true, nil
		}
		return spendCheck.FirstSeen.Before(spend.FirstSeen), nil
	}
	return false, jerr.Newf("error no spend found to compare against")
}

type DoubleSpendCheckSpend struct {
	TxHash    []byte
	Index     uint32
	FirstSeen time.Time
	BlockHash []byte
}
