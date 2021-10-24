package test_tx

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/ref/bitcoin/memo"
)

type TestGetter struct {
	UTXOs []memo.UTXO
}

func (g *TestGetter) SetPkHashesToUse([][]byte) {}

func (g *TestGetter) GetUTXOsOld(request *memo.UTXORequest) ([]memo.UTXO, error) {
	size := jutil.MinInt(25, len(g.UTXOs))
	if size == 0 {
		return nil, nil
	}
	toRet := g.UTXOs[:size]
	g.UTXOs = g.UTXOs[size:]
	return toRet, nil
}

type TestFaucetGetter struct {
	TestGetter
}

func (g *TestFaucetGetter) SetPkHashesToUse([][]byte) {}

func (g *TestFaucetGetter) GetUTXOsOld(request *memo.UTXORequest) ([]memo.UTXO, error) {
	utxos, err := g.TestGetter.GetUTXOsOld(request)
	if err != nil {
		return nil, jerr.Get("error getting utxos", err)
	}
	if len(utxos) > 1 {
		return nil, jerr.Newf("error invalid utxo count for faucet (%d)", len(utxos))
	}
	return utxos, nil
}
