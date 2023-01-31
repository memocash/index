package wlt

import (
	"fmt"
	"github.com/memocash/index/client/lib"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/tx/script"
	"github.com/memocash/index/ref/bitcoin/wallet"
)

type InputGetter struct {
	Address wallet.Addr
	UTXOs   []memo.UTXO
	Client  *lib.Client
	reset   bool
}

func NewInputGetter(address wallet.Addr, client *lib.Client) *InputGetter {
	return &InputGetter{
		Address: address,
		Client:  client,
	}
}

func (g *InputGetter) SetPkHashesToUse([][]byte) {
}

func (g *InputGetter) GetUTXOs(*memo.UTXORequest) ([]memo.UTXO, error) {
	if g.reset && len(g.UTXOs) > 0 {
		g.reset = false
		return g.UTXOs, nil
	}
	outputs, err := g.Client.GetUtxos(g.Address)
	if err != nil {
		return nil, fmt.Errorf("error getting utxos from input getter client; %w", err)
	}
	var utxos []memo.UTXO
	pkHash := g.Address.GetPkHash()
	pkScript, err := script.P2pkh{PkHash: pkHash}.Get()
	if err != nil {
		return nil, fmt.Errorf("error getting pk script; %w", err)
	}
	for _, output := range outputs {
		utxos = append(utxos, memo.UTXO{
			Input: memo.TxInput{
				PkScript:     pkScript,
				PkHash:       pkHash,
				Value:        output.Amount,
				PrevOutHash:  hs.GetTxHash(output.Hash),
				PrevOutIndex: uint32(output.Index),
			},
		})
	}
	g.UTXOs = utxos
	return utxos, nil
}

func (g *InputGetter) MarkUTXOsUsed(used []memo.UTXO) {
	for i := 0; i < len(g.UTXOs); i++ {
		for j := 0; j < len(used); j++ {
			if g.UTXOs[i].IsEqual(used[j]) {
				g.UTXOs = append(g.UTXOs[:i], g.UTXOs[i+1:]...)
				i--
				break
			}
		}
	}
}

func (g *InputGetter) AddChangeUTXO(utxo memo.UTXO) {
	g.UTXOs = append(g.UTXOs, utxo)
}

func (g *InputGetter) NewTx() {
	g.reset = true
}
