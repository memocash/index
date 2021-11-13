package test_block

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"math/rand"
	"time"
)

var _defaultGenerator *BlockGenerator

func ResetDefaultGenerator() {
	_defaultGenerator = new(BlockGenerator)
}

func init() {
	ResetDefaultGenerator()
}

func GetNextBlock(txs []*wire.MsgTx) *wire.MsgBlock {
	return _defaultGenerator.GetNextBlock(txs)
}

type BlockGenerator struct {
	PrevBlock chainhash.Hash
	Time      time.Time
}

func (g *BlockGenerator) GetNextBlock(txs []*wire.MsgTx) *wire.MsgBlock {
	if g.Time.IsZero() {
		g.Time = time.Date(2009, 1, 3, 18, 15, 5, 0, time.UTC)
	} else {
		g.Time = g.Time.Add(GetRandom10Minute())
	}
	var block = &wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    1,
			PrevBlock:  g.PrevBlock,
			MerkleRoot: chainhash.Hash{},
			Timestamp:  g.Time,
		},
		Transactions: txs,
	}
	for {
		g.PrevBlock = block.BlockHash()
		if g.PrevBlock.String()[:3] == "000" {
			break
		}
		block.Header.Nonce++
	}
	return block
}

func GetRandom10Minute() time.Duration {
	const maxDuration = 10 * time.Minute
	const rolls = 4
	var duration time.Duration
	for i := 0; i < rolls; i++ {
		duration += time.Duration(rand.Int63n(int64(maxDuration) / rolls))
	}
	return duration
}
