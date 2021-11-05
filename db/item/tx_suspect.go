package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/server/db/client"
)

type TxSuspect struct {
	TxHash []byte
}

func (s TxSuspect) GetUid() []byte {
	return jutil.ByteReverse(s.TxHash)
}

func (s TxSuspect) GetShard() uint {
	return client.GetByteShard(s.TxHash)
}

func (s TxSuspect) GetTopic() string {
	return TopicTxSuspect
}

func (s TxSuspect) Serialize() []byte {
	return nil
}

func (s *TxSuspect) SetUid(uid []byte) {
	if len(uid) != 32 {
		return
	}
	s.TxHash = jutil.ByteReverse(uid)
}

func (s *TxSuspect) Deserialize([]byte) {}
