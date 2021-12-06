package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/config"
)

type HeightProcessed struct {
	Name   string
	Height int64
	Block  []byte
}

func (p HeightProcessed) GetUid() []byte {
	return []byte(p.Name)
}

func (p HeightProcessed) GetShard() uint {
	return GetShardByte(p.GetUid())
}

func (p HeightProcessed) GetTopic() string {
	return TopicHeightProcessed
}

func (p HeightProcessed) Serialize() []byte {
	return jutil.CombineBytes(jutil.GetInt64DataBig(p.Height), jutil.ByteReverse(p.Block))
}

func (p *HeightProcessed) SetUid(uid []byte) {
	p.Name = string(uid)
}

func (p *HeightProcessed) Deserialize(data []byte) {
	if len(data) != 40 {
		return
	}
	p.Height = jutil.GetInt64Big(data[:8])
	p.Block = jutil.ByteReverse(data[8:])
}

func (p *HeightProcessed) Save() error {
	var objs = []Object{p}
	err := Save(objs)
	if err != nil {
		return jerr.Get("error saving height processed", err)
	}
	return nil
}

func GetRecentHeightProcessed(name string) (*HeightProcessed, error) {
	cfg := config.GetShardConfig(GetShardByte32([]byte(name)), config.GetQueueShards())
	db := client.NewClient(cfg.GetHost())
	if err := db.GetSingle(TopicHeightProcessed, []byte(name)); err != nil {
		return nil, jerr.Getf(err, "error getting recent height block")
	} else if len(db.Messages) > 1 {
		return nil, jerr.Newf("error unexpected number of height processed messages: %d", len(db.Messages))
	}
	var heightProcessed = new(HeightProcessed)
	heightProcessed.SetUid(db.Messages[0].Uid)
	heightProcessed.Deserialize(db.Messages[0].Message)
	return heightProcessed, nil
}
