package slp

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type Send struct {
	TxHash    [32]byte
	TokenHash [32]byte
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
	return jutil.ByteReverse(s.TokenHash[:])
}

func (s *Send) Deserialize(data []byte) {
	if len(data) != memo.TxHashLength {
		return
	}
	copy(s.TokenHash[:], jutil.ByteReverse(data))
}
