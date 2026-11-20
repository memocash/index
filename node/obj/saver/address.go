package saver

import (
	"context"
	"fmt"
	"log"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/db/item/addr"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"github.com/memocash/index/ref/dbi"
)

// Address writes the addr_seen_tx rows for every address a tx pays to or
// spends from. Outputs resolve to any supported lock script kind, P2SH32
// included. Inputs resolve from the unlock script alone: P2PKH to its pubkey
// hash, and a P2SH-shaped spend to the P2SH20 alias of its redeem script,
// whichever hash the spent output actually demanded (the saver never reads
// the spent output, so it cannot know). Each P2SH-shaped spend also writes
// the P2SH20/P2SH32 link for its redeem script, so readers of a P2SH32
// address find its spends through the alias.
type Address struct {
	Verbose     bool
	P2shCount   int64
	P2sh32Count int64
	P2pkhCount  int64
	SkipP2pkh   bool
}

func (a *Address) SaveTxs(ctx context.Context, b *dbi.Block) error {
	if b.IsNil() {
		return fmt.Errorf("error nil block")
	}
	var objects []db.Object
	var links = make(map[wallet.Addr32]wallet.Addr)
	for _, transaction := range b.Transactions {
		var tx = transaction.MsgTx
		txHash := chainhash.Hash(transaction.Hash)
		if a.Verbose {
			log.Printf("tx: %s\n", txHash.String())
		}
		var addrs = make(map[wallet.LockAddr]struct{})
		for j := range tx.TxIn {
			if memo.IsCoinbaseInput(tx.TxIn[j]) {
				continue
			}
			if address, err := wallet.GetP2pkhAddrFromUnlockScript(tx.TxIn[j].SignatureScript); err == nil {
				addrs[address.LockAddr()] = struct{}{}
				continue
			}
			redeemScript, err := wallet.GetP2shRedeemScriptFromUnlockScript(tx.TxIn[j].SignatureScript)
			if err != nil {
				continue
			}
			p2shAddr, p2sh32Addr := wallet.GetP2shAddrsFromRedeemScript(redeemScript)
			addrs[p2shAddr.LockAddr()] = struct{}{}
			links[p2sh32Addr] = p2shAddr
		}
		for h := range tx.TxOut {
			address, err := wallet.GetLockAddrFromLockScript(tx.TxOut[h].PkScript)
			if err != nil {
				continue
			}
			addrs[address] = struct{}{}
		}
		for address := range addrs {
			switch {
			case address.IsP2SH32():
				a.P2sh32Count++
				if a.Verbose {
					log.Printf("p2sh32 output: %s (%s)\n", address.String(), txHash.String())
				}
			case address.Addr().IsP2SH():
				a.P2shCount++
				if a.Verbose {
					log.Printf("p2sh output: %s (%s)\n", address.String(), txHash.String())
				}
			case address.Addr().IsP2PKH():
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
	for p2sh32Addr, p2shAddr := range links {
		objects = append(objects, addr.NewP2shLinks(p2shAddr, p2sh32Addr)...)
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
