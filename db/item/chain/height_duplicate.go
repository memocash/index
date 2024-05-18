package chain

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/config"
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

func GetHeightDuplicatesAll(startHeight int64) ([]*HeightDuplicate, error) {
	var heightDuplicates []*HeightDuplicate
	for _, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.GetLarge(db.TopicChainHeightDuplicate, jutil.GetInt64DataBig(startHeight), false, false)
		if err != nil {
			return nil, jerr.Get("error getting height duplicates from queue client", err)
		}
		for i := range dbClient.Messages {
			var heightDuplicate = new(HeightDuplicate)
			db.Set(heightDuplicate, dbClient.Messages[i])
			heightDuplicates = append(heightDuplicates, heightDuplicate)
		}
	}
	return heightDuplicates, nil
}
