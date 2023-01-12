package saver

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/index/db/item"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
)

type Address struct {
	Verbose     bool
	InitialSync bool
	P2shCount   int64
	P2pkhCount  int64
}

func (a *Address) SaveTxs(b *dbi.Block) error {
	if b.IsNil() {
		return jerr.Newf("error nil block")
	}
	block := b.ToWireBlock()
	var height = b.Height
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
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			address, err := wallet.GetAddrFromUnlockScript(tx.TxIn[j].SignatureScript)
			if err != nil {
				//jerr.Get("error getting address from unlock script", err).Print()
				continue
			}
			heightInput := &addr.HeightInput{
				Addr:   *address,
				Height: int32(height),
				TxHash: txHash,
				Index:  uint32(j),
			}
			if address.IsP2PKH() {
				a.P2pkhCount++
			} else if address.IsP2SH() {
				a.P2shCount++
				if a.Verbose {
					jlog.Logf("p2sh input: %s (%s)\n", address.String(), txHash.String())
				}
			}
			objects = append(objects, heightInput)
			if !a.InitialSync && height != item.HeightMempool {
				objectsToRemove = append(objectsToRemove, &addr.HeightInput{
					Addr:   heightInput.Addr,
					Height: item.HeightMempool,
					TxHash: heightInput.TxHash,
					Index:  heightInput.Index,
				})
			}
		}
		for h := range tx.TxOut {
			address, err := wallet.GetAddrFromLockScript(tx.TxOut[h].PkScript)
			if err != nil {
				continue
			}
			heightOutput := &addr.HeightOutput{
				Addr:   *address,
				Height: int32(height),
				TxHash: txHash,
				Index:  uint32(h),
				Value:  tx.TxOut[h].Value,
			}
			if address.IsP2PKH() {
				a.P2pkhCount++
			} else if address.IsP2SH() {
				a.P2shCount++
				if a.Verbose {
					jlog.Logf("p2sh output: %s (%s)\n", address.String(), txHash.String())
				}
			}
			objects = append(objects, heightOutput)
			if !a.InitialSync && height != item.HeightMempool {
				objectsToRemove = append(objectsToRemove, &addr.HeightOutput{
					Addr:   heightOutput.Addr,
					Height: item.HeightMempool,
					TxHash: heightOutput.TxHash,
					Index:  heightOutput.Index,
					Value:  heightOutput.Value,
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

func NewAddress(verbose, initial bool) *Address {
	return &Address{
		Verbose:     verbose,
		InitialSync: initial,
	}
}
