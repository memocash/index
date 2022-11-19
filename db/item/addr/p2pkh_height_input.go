package addr

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type P2pkhHeightInput struct {
	PkHash [20]byte
	Height int32
	TxHash [32]byte
	Index  uint32
}

func (i *P2pkhHeightInput) GetTopic() string {
	return db.TopicP2pkhHeightInput
}

func (i *P2pkhHeightInput) GetShard() uint {
	return client.GetByteShard(i.PkHash[:])
}

func (i *P2pkhHeightInput) GetUid() []byte {
	return GetPkHashHeightTxHashIndexUid(i.PkHash, i.Height, i.TxHash, i.Index)
}

func (i *P2pkhHeightInput) SetUid(uid []byte) {
	if len(uid) != 60 {
		return
	}
	copy(i.PkHash[:], uid[:20])
	i.Height = jutil.GetInt32Big(uid[20:24])
	copy(i.TxHash[:], jutil.ByteReverse(uid[24:56]))
	i.Index = jutil.GetUint32Big(uid[56:60])
}

func (i *P2pkhHeightInput) Serialize() []byte {
	return nil
}

func (i *P2pkhHeightInput) Deserialize([]byte) {}
