package slp

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Mint struct {
	TxHash     [32]byte
	TokenHash  [32]byte
	TokenType  uint8
	BatonIndex uint32
	Quantity   uint64
}

func (m *Mint) GetTopic() string {
	return db.TopicSlpMint
}

func (m *Mint) GetShardSource() uint {
	return client.GenShardSource(m.TxHash[:])
}

func (m *Mint) GetUid() []byte {
	return jutil.ByteReverse(m.TxHash[:])
}

func (m *Mint) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(m.TxHash[:], jutil.ByteReverse(uid))
}

func (m *Mint) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(m.TokenHash[:]),
		jutil.GetUint32Data(m.BatonIndex),
		jutil.GetUint64Data(m.Quantity),
		[]byte{m.TokenType},
	)
}

func (m *Mint) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength+4+8+1 {
		return
	}
	copy(m.TokenHash[:], jutil.ByteReverse(data[:memo.TxHashLength]))
	m.BatonIndex = jutil.GetUint32(data[memo.TxHashLength : memo.TxHashLength+4])
	m.Quantity = jutil.GetUint64(data[memo.TxHashLength+4 : memo.TxHashLength+12])
	m.TokenType = data[memo.TxHashLength+12]
}

func GetMints(ctx context.Context, txHashes [][32]byte) ([]*Mint, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardIdFromByte32(txHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(txHash[:]))
	}
	messages, err := db.GetSpecific(ctx, db.TopicSlpMint, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting slp mints; %w", err)
	}
	var mints []*Mint
	for i := range messages {
		var mint = new(Mint)
		db.Set(mint, messages[i])
		mints = append(mints, mint)
	}
	return mints, nil
}
