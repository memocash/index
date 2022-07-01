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
		jutil.ByteReverse(n.TxHash),
	)
}

func (n MemoName) GetShard() uint {
	return client.GetByteShard(n.LockHash)
}

func (n MemoName) GetTopic() string {
	return TopicMemoName
}

func (n MemoName) Serialize() []byte {
	return []byte(n.Name)
}

func (n *MemoName) SetUid(uid []byte) {
	if len(uid) != memo.TxHashLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	n.LockHash = uid[:32]
	n.Height = jutil.GetInt64Big(uid[32:40])
	n.TxHash = jutil.ByteReverse(uid[40:72])
}

func (n *MemoName) Deserialize(data []byte) {
	n.Name = string(data)
}
