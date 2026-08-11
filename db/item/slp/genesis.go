package slp

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Genesis struct {
	TxHash     [32]byte
	TokenType  uint8
	Decimals   uint8
	BatonIndex uint32
	Ticker     string
	Name       string
	DocUrl     string
	DocHash    [32]byte
}

func (g *Genesis) GetTopic() string {
	return db.TopicSlpGenesis
}

func (g *Genesis) GetShardSource() uint {
	return client.GenShardSource(g.TxHash[:])
}

func (g *Genesis) GetUid() []byte {
	return jutil.ByteReverse(g.TxHash[:])
}

func (g *Genesis) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(g.TxHash[:], jutil.ByteReverse(uid))
}

// genesisFixedLen is the fixed prefix of a serialized genesis: token type,
// decimals, baton index, and doc hash. The ticker, name, and doc url follow,
// each as a uint16 length prefix and the raw bytes (on-chain fields are
// arbitrary bytes, so they are stored byte-faithfully).
const genesisFixedLen = 2 + 4 + memo.TxHashLength

func (g *Genesis) Serialize() []byte {
	return jutil.CombineBytes(
		[]byte{g.TokenType, g.Decimals},
		jutil.GetUint32Data(g.BatonIndex),
		g.DocHash[:],
		jutil.GetUint16Data(uint16(len(g.Ticker))), []byte(g.Ticker),
		jutil.GetUint16Data(uint16(len(g.Name))), []byte(g.Name),
		jutil.GetUint16Data(uint16(len(g.DocUrl))), []byte(g.DocUrl),
	)
}

func (g *Genesis) Deserialize(data []byte) {
	var fields [3]string
	var offset = genesisFixedLen
	for i := range fields {
		if len(data) < offset+2 {
			return
		}
		fieldLen := int(jutil.GetUint16(data[offset : offset+2]))
		offset += 2
		if len(data) < offset+fieldLen {
			return
		}
		fields[i] = string(data[offset : offset+fieldLen])
		offset += fieldLen
	}
	if offset != len(data) {
		return
	}
	g.TokenType = data[0]
	g.Decimals = data[1]
	g.BatonIndex = jutil.GetUint32(data[2:6])
	copy(g.DocHash[:], data[6:genesisFixedLen])
	g.Ticker = fields[0]
	g.Name = fields[1]
	g.DocUrl = fields[2]
}

func GetGeneses(ctx context.Context, txHashes [][32]byte) ([]*Genesis, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardIdFromByte32(txHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(txHash[:]))
	}
	messages, err := db.GetSpecific(ctx, db.TopicSlpGenesis, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting slp geneses; %w", err)
	}
	var geneses []*Genesis
	for i := range messages {
		var genesis = new(Genesis)
		db.Set(genesis, messages[i])
		geneses = append(geneses, genesis)
	}
	return geneses, nil
}
