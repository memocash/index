package saver

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
)

type Address struct {
	Verbose     bool
	InitialSync bool
}

func (a *Address) SaveTxs(b *dbi.Block) error {
	if b.IsNil() {
		return jerr.Newf("error nil block")
	}
	block := b.ToWireBlock()
	var height int64
	if !block.Header.Timestamp.IsZero() {
		blockHash := block.BlockHash()
		blockHashBytes := blockHash.CloneBytes()
		blockHeight, err := chain.GetBlockHeight(blockHashBytes)
		if err != nil {
			return jerr.Get("error getting block height for memo", err)
		}
		height = blockHeight.Height
	}
	if height == 0 {
		height = item.HeightMempool
	}
	var objects []db.Object
	var objectsToRemove []db.Object
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		if a.Verbose {
			jlog.Logf("tx: %s\n", txHash.String())
		}
		for j := range tx.TxIn {
			address := GetP2pkhAddressFromUnlockScript(tx.TxIn[j].SignatureScript)
			if !address.IsSet() {
				continue
			}
			p2pkhHeightInput := &addr.P2pkhHeightInput{
				Height: int32(height),
				TxHash: txHash,
				Index:  uint32(j),
			}
			copy(p2pkhHeightInput.PkHash[:], address.GetPkHash())
			objects = append(objects, p2pkhHeightInput)
			if !a.InitialSync {
				objectsToRemove = append(objectsToRemove, &addr.P2pkhHeightInput{
					PkHash: p2pkhHeightInput.PkHash,
					Height: item.HeightMempool,
					TxHash: p2pkhHeightInput.TxHash,
					Index:  p2pkhHeightInput.Index,
				})
			}
		}
		for h := range tx.TxOut {
			address := GetP2pkhAddressFromLockScript(tx.TxOut[h].PkScript)
			if !address.IsSet() {
				continue
			}
			p2pkhHeightOutput := &addr.P2pkhHeightOutput{
				Height: int32(height),
				TxHash: txHash,
				Index:  uint32(h),
			}
			copy(p2pkhHeightOutput.PkHash[:], address.GetPkHash())
			objects = append(objects, p2pkhHeightOutput)
			if !a.InitialSync {
				objectsToRemove = append(objectsToRemove, &addr.P2pkhHeightOutput{
					PkHash: p2pkhHeightOutput.PkHash,
					Height: item.HeightMempool,
					TxHash: p2pkhHeightOutput.TxHash,
					Index:  p2pkhHeightOutput.Index,
				})
			}
		}
	}
	if err := db.Save(objects); err != nil {
		return jerr.Get("error saving db tx objects", err)
	}
	if a.InitialSync {
		return nil
	}
	if err := db.Remove(objectsToRemove); err != nil {
		return jerr.Get("error removing mempool lock height outputs for lock heights", err)
	}
	return nil
}

func NewAddress(verbose bool) *Address {
	return &Address{
		Verbose: verbose,
	}
}

func GetP2pkhAddressFromUnlockScript(unlockScript []byte) wallet.Address {
	l := len(unlockScript)
	if l < 2 || !jutil.InIntSlice(int(unlockScript[0]),
		[]int{txscript.OP_DATA_64, txscript.OP_DATA_65, txscript.OP_DATA_71, txscript.OP_DATA_72}) {
		return wallet.Address{}
	}
	s := int(unlockScript[0])
	if l < s+35 || unlockScript[s+1] != txscript.OP_DATA_33 {
		return wallet.Address{}
	}
	return wallet.GetAddress(unlockScript[s+2:])
}

func GetP2pkhAddressFromLockScript(lockScript []byte) wallet.Address {
	if len(lockScript) != 25 || lockScript[0] != txscript.OP_DUP || lockScript[1] != txscript.OP_HASH160 ||
		lockScript[2] != txscript.OP_DATA_20 || lockScript[23] != txscript.OP_EQUALVERIFY ||
		lockScript[24] != txscript.OP_CHECKSIG {
		return wallet.Address{}
	}
	return wallet.GetAddressFromPkHash(lockScript[3:23])
}
