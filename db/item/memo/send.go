package memo

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
	Recipient [25]byte
}

func (s *Send) GetTopic() string {
	return db.TopicMemoSend
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
	return s.Recipient[:]
}

func (s *Send) Deserialize(data []byte) {
	if len(data) != memo.AddressLength {
		return
	}
	copy(s.Recipient[:], data)
}

func GetSends(ctx context.Context, txHashes [][32]byte) ([]*Send, error) {
	messages, err := db.GetSpecific(ctx, db.TopicMemoSend, db.ShardUidsTxHashes(txHashes))
	if err != nil {
		return nil, fmt.Errorf("error getting memo sends; %w", err)
	}
	sends := make([]*Send, len(messages))
	for i := range messages {
		sends[i] = new(Send)
		db.Set(sends[i], messages[i])
	}
	return sends, nil
}
