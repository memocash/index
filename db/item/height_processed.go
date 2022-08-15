package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
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
	return db.GetShardByte(p.GetUid())
}

func (p HeightProcessed) GetTopic() string {
	return db.TopicHeightProcessed
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
	var objs = []db.Object{p}
	if err := db.Save(objs); err != nil {
		return jerr.Get("error saving height processed", err)
	}
	return nil
}

func GetRecentHeightProcessed(name string) (*HeightProcessed, error) {
	cfg := config.GetShardConfig(db.GetShardByte32([]byte(name)), config.GetQueueShards())
	dbClient := client.NewClient(cfg.GetHost())
	if err := dbClient.GetSingle(db.TopicHeightProcessed, []byte(name)); err != nil {
		return nil, jerr.Getf(err, "error getting recent height block")
	} else if len(dbClient.Messages) > 1 {
		return nil, jerr.Newf("error unexpected number of height processed messages: %d", len(dbClient.Messages))
	}
	var heightProcessed = new(HeightProcessed)
	db.Set(heightProcessed, dbClient.Messages[0])
	return heightProcessed, nil
}
