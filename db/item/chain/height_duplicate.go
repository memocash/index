package chain

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/db"
)

type HeightDuplicate struct {
	Height    int64
	BlockHash [32]byte
}

func (d *HeightDuplicate) GetTopic() string {
	return db.TopicChainHeightDuplicate
}

func (d *HeightDuplicate) GetShardSource() uint {
	return uint(d.Height)
}

func (d *HeightDuplicate) GetUid() []byte {
	return jutil.CombineBytes(jutil.GetInt64DataBig(d.Height), jutil.ByteReverse(d.BlockHash[:]))
}

func (d *HeightDuplicate) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	d.Height = jutil.GetInt64Big(uid[:8])
	copy(d.BlockHash[:], jutil.ByteReverse(uid[8:40]))
}

func (d *HeightDuplicate) Serialize() []byte {
	return nil
}

func (d *HeightDuplicate) Deserialize([]byte) {}
