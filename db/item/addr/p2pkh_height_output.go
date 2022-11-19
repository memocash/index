package addr

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
)

type P2pkhHeightOutput struct {
	PkHash [20]byte
	Height int32
	TxHash [32]byte
	Index  uint32
}

func (o *P2pkhHeightOutput) GetTopic() string {
	return db.TopicP2pkhHeightOutput
}

func (o *P2pkhHeightOutput) GetShard() uint {
	return client.GetByteShard(o.PkHash[:])
}

func (o *P2pkhHeightOutput) GetUid() []byte {
	return GetPkHashHeightTxHashIndexUid(o.PkHash, o.Height, o.TxHash, o.Index)
}

func (o *P2pkhHeightOutput) SetUid(uid []byte) {
	if len(uid) != 60 {
		return
	}
	copy(o.PkHash[:], uid[:20])
	o.Height = jutil.GetInt32Big(uid[20:24])
	copy(o.TxHash[:], jutil.ByteReverse(uid[24:56]))
	o.Index = jutil.GetUint32Big(uid[56:60])
}

func (o *P2pkhHeightOutput) Serialize() []byte {
	return nil
}

func (o *P2pkhHeightOutput) Deserialize([]byte) {}

func GetPkHashHeightTxHashIndexUid(pkHash [20]byte, height int32, txHash [32]byte, index uint32) []byte {
	return jutil.CombineBytes(
		pkHash[:],
		jutil.GetInt32DataBig(height),
		jutil.ByteReverse(txHash[:]),
		jutil.GetUint32DataBig(index),
	)
}
