package saver

import (
	"bytes"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type Memo struct {
	Verbose bool
	Shard   int
}

func (t *Memo) SaveTxs(block *wire.MsgBlock) error {
	if block == nil {
		return jerr.Newf("error nil block")
	}
	var height int64
	if !block.Header.Timestamp.IsZero() {
		blockHash := block.BlockHash()
		blockHashBytes := blockHash.CloneBytes()
		blockHeight, err := item.GetBlockHeight(blockHashBytes)
		if err != nil {
			return jerr.Get("error getting block height for memo", err)
		}
		height = blockHeight.Height
	}
	if height == 0 {
		height = item.HeightMempool
	}
	memoCodes := [][]byte{
		memo.PrefixSetName,
	}
	var objects []item.Object
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		txHashBytes := txHash.CloneBytes()
		if t.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		var setNames []*item.MemoName
		for _, memoCode := range memoCodes {
			memoScript, err := memo.GetBaseOpReturn().AddData(memoCode).Script()
			if err != nil {

			}
			for h := range tx.TxOut {
				if len(tx.TxOut[h].PkScript) < len(memoScript) ||
					!bytes.Equal(tx.TxOut[h].PkScript[:len(memoScript)], memoScript) {
					continue
				}
				pushData, err := txscript.PushedData(tx.TxOut[h].PkScript)
				if err != nil {
					return jerr.Get("error getting pushed data", err)
				}
				if len(pushData) != 2 {
					return jerr.Newf("invalid set name, incorrect push data (%d)", len(pushData))
				}
				var name = jutil.GetUtf8String(pushData[1])
				setNames = append(setNames, &item.MemoName{
					LockHash: nil,
					Height:   height,
					TxHash:   txHashBytes,
					Name:     name,
				})
			}
		}
		var inputPkHash []byte
		for j := range tx.TxIn {
			address, err := wallet.GetAddressFromSignatureScript(tx.TxIn[j].SignatureScript)
			if err != nil {
				return jerr.Get("error getting address from signature script", err)
			}
			inputPkHash = address.GetPkHash()
			if len(inputPkHash) > 0 {
				break
			}
		}
		if len(inputPkHash) == 0 {
			return jerr.New("error could not find input pk hash for memo")
		}
		for _, setName := range setNames {
			setName.LockHash = inputPkHash
			objects = append(objects, setName)
		}
	}
	if err := item.Save(objects); err != nil {
		return jerr.Get("error saving db memo objects", err)
	}
	return nil
}

func NewMemoShard(verbose bool, shard int) *Memo {
	return &Memo{
		Verbose: verbose,
		Shard:   shard,
	}
}
