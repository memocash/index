package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/memocash/index/client/lib"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"golang.org/x/term"
	"syscall"
)

type Wallet struct {
	Client  *lib.Client
	Utxos   []memo.UTXO
	Used    bool
	Change  wallet.Change
	KeyRing wallet.KeyRing
}

func NewWalletFromStdinWif() (*Wallet, error) {
	fmt.Printf("Enter WIF: ")
	wif, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, fmt.Errorf("error reading wif from stdin; %w", err)
	}
	fmt.Println()
	wlt, err := NewWallet(string(wif))
	if err != nil {
		return nil, fmt.Errorf("error creating new wallet from stdin wif; %w", err)
	}
	return wlt, nil
}

func NewWallet(wif string) (*Wallet, error) {
	privateKey, err := wallet.ImportPrivateKey(wif)
	if err != nil {
		return nil, fmt.Errorf("error getting private key; %w", err)
	}
	client, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("error getting client; %w", err)
	}
	clientUtxos, err := client.GetUtxos([]wallet.Addr{privateKey.GetAddr()})
	if err != nil {
		return nil, fmt.Errorf("error getting client utxos; %w", err)
	}
	var memoUtxos []memo.UTXO
	for _, utxo := range clientUtxos {
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
		Client:  client,
		Utxos:   memoUtxos,
		Change:  wallet.Change{Main: privateKey.GetAddress()},
		KeyRing: wallet.KeyRing{Keys: []wallet.PrivateKey{privateKey}},
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
