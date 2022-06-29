package item

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/memo"
)

type MemoName struct {
	LockHash []byte
	Height   int64
	TxHash   []byte
	Name     string
}

func (n MemoName) GetUid() []byte {
	return jutil.CombineBytes(
		n.LockHash,
		jutil.GetInt64DataBig(n.Height),
	)
}

func (n MemoName) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoName) GetTopic() string {
	return TopicMemoName
}

func (n MemoName) Serialize() []byte {
	return jutil.CombineBytes(
		jutil.ByteReverse(n.TxHash),
		[]byte(n.Name),
	)
}

func (n *MemoName) SetUid(uid []byte) {
	if len(uid) != 68 {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(uid[32:])
}

func (n *MemoName) Deserialize(data []byte) {
	if len(data) < memo.TxHashLength {
		return
	}
	n.TxHash = jutil.ByteReverse(data[0:32])
	n.Name = string(data[32:])
}
