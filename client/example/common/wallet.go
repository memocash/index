package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
)

type Wallet struct {
	Utxos  []memo.UTXO
	Used   bool
	Change wallet.Change
}

func NewWallet(addressStrings []string) (*Wallet, error) {
	var addresses []wallet.Addr
	for _, addressString := range addressStrings {
		address, err := wallet.GetAddrFromString(addressString)
		if err != nil {
			log.Fatalf("error getting address from string; %v", err)
		}
		addresses = append(addresses, *address)
	}
	client, err := GetClient()
	if err != nil {
		log.Fatalf("error getting client; %v", err)
	}
	graphUtxos, err := client.GetUtxos(addresses)
	if err != nil {
		log.Fatalf("error getting utxos; %v", err)
	}
	var memoUtxos []memo.UTXO
	for _, utxo := range graphUtxos {
		txHash, err := chainhash.NewHashFromStr(utxo.Hash)
		if err != nil {
			return nil, fmt.Errorf("error getting hash from string for wallet utxos; %w", err)
		}
		script, err := hex.DecodeString(utxo.Script)
		if err != nil {
			return nil, fmt.Errorf("error decoding script for wallet utxos; %w", err)
		}
		addr, err := wallet.GetAddrFromString(utxo.Lock.Address)
		if err != nil {
			return nil, fmt.Errorf("error getting address from string for wallet utxos; %w", err)
		}
		memoUtxos = append(memoUtxos, memo.UTXO{Input: memo.TxInput{
			PkScript:     script,
			PkHash:       addr.GetPkHash(),
			PrevOutHash:  txHash[:],
			PrevOutIndex: uint32(utxo.Index),
			Value:        utxo.Amount,
		}})
	}
	return &Wallet{
		Utxos:  memoUtxos,
		Change: wallet.Change{Main: addresses[0].OldAddress()},
	}, nil
}

func (w *Wallet) SetPkHashesToUse([][]byte) {}
func (w *Wallet) GetUTXOs(*memo.UTXORequest) ([]memo.UTXO, error) {
	if w.Used {
		return nil, nil
	}
	w.Used = true
	return w.Utxos, nil
}

func (w *Wallet) MarkUTXOsUsed(used []memo.UTXO) {
	for _, u := range used {
		for i := 0; i < len(w.Utxos); i++ {
			if bytes.Equal(u.Input.PrevOutHash, w.Utxos[i].Input.PrevOutHash) &&
				u.Input.PrevOutIndex == w.Utxos[i].Input.PrevOutIndex {
				w.Utxos = append(w.Utxos[:i], w.Utxos[i+1:]...)
				i--
			}
		}
	}
}

func (w *Wallet) AddChangeUTXO(change memo.UTXO) {
	w.Utxos = append(w.Utxos, change)
}

func (w *Wallet) NewTx() {
	w.Used = false
}
