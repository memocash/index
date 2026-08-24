package slp

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Send struct {
	TxHash    [32]byte
	TokenHash [32]byte
	TokenType uint8
}

func (s *Send) GetTopic() string {
	return db.TopicSlpSend
}

func (s *Send) GetShardSource() uint {
	return client.GenShardSource(s.TxHash[:])
}

func (s *Send) GetUid() []byte {
	return jutil.ByteReverse(s.TxHash[:])
}

func (s *Send) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength {
		return
	}
	copy(s.TxHash[:], jutil.ByteReverse(uid))
}

func (s *Send) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(s.TokenHash[:]),
		[]byte{s.TokenType},
	)
}

func (s *Send) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength+1 {
		return
	}
	copy(s.TokenHash[:], jutil.ByteReverse(data[:memo.TxHashLength]))
	s.TokenType = data[memo.TxHashLength]
}

func GetSends(ctx context.Context, txHashes [][32]byte) ([]*Send, error) {
	var shardUids = make(map[uint32][][]byte)
	for _, txHash := range txHashes {
		shard := db.GetShardIdFromByte32(txHash[:])
		shardUids[shard] = append(shardUids[shard], jutil.ByteReverse(txHash[:]))
	}
	messages, err := db.GetSpecific(ctx, db.TopicSlpSend, shardUids)
	if err != nil {
		return nil, fmt.Errorf("error getting slp sends; %w", err)
	}
	var sends []*Send
	for i := range messages {
		var send = new(Send)
		db.Set(send, messages[i])
		sends = append(sends, send)
	}
	return sends, nil
}
