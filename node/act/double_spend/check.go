package double_spend

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/node/act/tx_raw"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"time"
)

type DoubleSpendCheck struct {
	ParentTxHash  []byte
	ParentTxIndex uint32
	LockHash      []byte
	Spends        []*DoubleSpendCheckSpend
}

func (c DoubleSpendCheck) IsWinnerSpend(spendCheck *DoubleSpendCheckSpend) (bool, error) {
	winnerSpend, err := c.GetWinnerSpend()
	if err != nil {
		return false, jerr.Get("error getting winner spend", err)
	}
	return bytes.Equal(winnerSpend.TxHash, spendCheck.TxHash), nil
}

func (c DoubleSpendCheck) GetWinnerSpend() (*DoubleSpendCheckSpend, error) {
	var winnerSpend *DoubleSpendCheckSpend
	for _, spend := range c.Spends {
		if winnerSpend == nil {
			winnerSpend = spend
			continue
		}
		if len(spend.BlockHash) > 0 && len(winnerSpend.BlockHash) == 0 {
			winnerSpend = spend
			continue
		}
		if len(spend.BlockHash) == 0 && len(winnerSpend.BlockHash) > 0 {
			continue
		}
		if !spend.FirstSeen.IsZero() && spend.FirstSeen.Before(winnerSpend.FirstSeen) {
			winnerSpend = spend
		}
	}
	return winnerSpend, nil
}

type DoubleSpendCheckSpend struct {
	TxHash     []byte
	Index      uint32
	LockHashes [][]byte
	FirstSeen  time.Time
	BlockHash  []byte
}

func AttachAllToDoubleSpendChecks(doubleSpendChecks []*DoubleSpendCheck) error {
	if err := AttachSeensToSpendCheckSpends(doubleSpendChecks); err != nil {
		return jerr.Get("error attaching seens to spend check spends", err)
	}
	if err := AttachBlocksToSpendCheckSpends(doubleSpendChecks); err != nil {
		return jerr.Get("error attaching blocks to spend check spends", err)
	}
	if err := AttachLockHashesToSpendChecks(doubleSpendChecks); err != nil {
		return jerr.Get("error attaching lock hashes to spend check spends", err)
	}
	return nil
}

func AttachSeensToSpendCheckSpends(doubleSpendChecks []*DoubleSpendCheck) error {
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
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, spend := range doubleSpendCheck.Spends {
			for _, txSeen := range txSeens {
				if bytes.Equal(txSeen.TxHash, spend.TxHash) {
					spend.FirstSeen = txSeen.Timestamp
					break
				}
			}
		}
	}
	return nil
}

// AttachBlocksToSpendCheckSpends
// TODO: Handle block hash already set, also include confirmation count
func AttachBlocksToSpendCheckSpends(doubleSpendChecks []*DoubleSpendCheck) error {
	var txHashes [][]byte
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, spend := range doubleSpendCheck.Spends {
			if len(spend.BlockHash) == 0 {
				txHashes = append(txHashes, spend.TxHash)
			}
		}
	}
	txBlocks, err := item.GetTxBlocks(txHashes)
	if err != nil {
		return jerr.Get("error getting tx blocks for double spend check spends", err)
	}
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, spend := range doubleSpendCheck.Spends {
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

// AttachLockHashesToSpendChecks assumes blocks attached before
func AttachLockHashesToSpendChecks(doubleSpendChecks []*DoubleSpendCheck) error {
	var txHashes [][]byte
	for _, doubleSpendCheck := range doubleSpendChecks {
		txHashes = append(txHashes, doubleSpendCheck.ParentTxHash)
		for _, spend := range doubleSpendCheck.Spends {
			txHashes = append(txHashes, spend.TxHash)
		}
	}
	txRaws, err := tx_raw.Get(txHashes)
	if err != nil {
		return jerr.Get("error getting tx raws for lock hashes double spend checks", err)
	}
	for _, doubleSpendCheck := range doubleSpendChecks {
		for _, txRaw := range txRaws {
			if bytes.Equal(txRaw.Hash, doubleSpendCheck.ParentTxHash) {
				msgTx, err := memo.GetMsgFromRaw(txRaw.Raw)
				if err != nil {
					return jerr.Getf(err, "error parsing raw msg for double spend check: %s",
						hs.GetTxString(doubleSpendCheck.ParentTxHash))
				}
				doubleSpendCheck.LockHash = script.GetLockHash(msgTx.TxOut[doubleSpendCheck.ParentTxIndex].PkScript)
				break
			}
		}
		for _, spend := range doubleSpendCheck.Spends {
			for _, txRaw := range txRaws {
				if bytes.Equal(txRaw.Hash, spend.TxHash) {
					msgTx, err := memo.GetMsgFromRaw(txRaw.Raw)
					if err != nil {
						return jerr.Getf(err, "error parsing raw msg for double spend check spend tx: %s", hs.GetTxString(spend.TxHash))
					}
					for _, out := range msgTx.TxOut {
						spend.LockHashes = append(spend.LockHashes, script.GetLockHash(out.PkScript))
					}
					break
				}
			}
		}
	}
	return nil
}
