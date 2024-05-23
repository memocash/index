package saver

import (
	"context"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
	"log"
)

type Address struct {
	Verbose    bool
	P2shCount  int64
	P2pkhCount int64
	SkipP2pkh  bool
}

func (a *Address) SaveTxs(ctx context.Context, b *dbi.Block) error {
	if b.IsNil() {
		return fmt.Errorf("error nil block")
	}
	var objects []db.Object
	for _, transaction := range b.Transactions {
		var tx = transaction.MsgTx
		txHash := chainhash.Hash(transaction.Hash)
		if a.Verbose {
			log.Printf("tx: %s\n", txHash.String())
		}
		var addrs = make(map[wallet.Addr]struct{})
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			address, err := wallet.GetAddrFromUnlockScript(tx.TxIn[j].SignatureScript)
			if err != nil {
				//log.Printf("error getting address from unlock script; %v", err)
				continue
			}
			addrs[*address] = struct{}{}
		}
		for h := range tx.TxOut {
			address, err := wallet.GetAddrFromLockScript(tx.TxOut[h].PkScript)
			if err != nil {
				continue
			}
			addrs[*address] = struct{}{}
		}
		for address := range addrs {
			if address.IsP2SH() {
				a.P2shCount++
				if a.Verbose {
					log.Printf("p2sh output: %s (%s)\n", address.String(), txHash.String())
				}
			} else if address.IsP2PKH() {
				a.P2pkhCount++
				if a.SkipP2pkh {
					continue
				}
			}
			objects = append(objects, &addr.SeenTx{
				Addr:   address,
				Seen:   transaction.Seen,
				TxHash: txHash,
			})
		}
	}
	if err := db.Save(objects); err != nil {
		return fmt.Errorf("error saving db tx objects; %w", err)
	}
	return nil
}

func NewAddress(verbose bool) *Address {
	return &Address{
		Verbose: verbose,
	}
}
