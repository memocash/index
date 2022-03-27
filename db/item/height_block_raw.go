package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/config"
	"strings"
)

type HeightBlockRaw struct {
	Height    int64
	BlockHash []byte
}

func (b HeightBlockRaw) GetUid() []byte {
	return jutil.CombineBytes(jutil.GetInt64DataBig(b.Height), jutil.ByteReverse(b.BlockHash))
}

func (b HeightBlockRaw) GetShard() uint {
	return uint(b.Height)
}

func (b HeightBlockRaw) GetTopic() string {
	return TopicHeightBlockRaw
}

func (b HeightBlockRaw) Serialize() []byte {
	return nil
}

func (b *HeightBlockRaw) SetUid(uid []byte) {
	if len(uid) != 40 {
		return
	}
	b.Height = jutil.GetInt64Big(uid[:8])
	b.BlockHash = jutil.ByteReverse(uid[8:40])
}

func (b *HeightBlockRaw) Deserialize([]byte) {}

func GetRecentHeightBlockRaw() (*HeightBlockRaw, error) {
	var heightBlockRaws []*HeightBlockRaw
	for i, shardConfig := range config.GetQueueShards() {
		dbClient := client.NewClient(shardConfig.GetHost())
		err := dbClient.Get(TopicHeightBlockRaw, client.GetMaxStart(), false)
		if err != nil {
			return nil, jerr.Getf(err, "error getting recent height block raw for shard: %d", i)
		}
		for i := range dbClient.Messages {
			var heightBlockRaw = new(HeightBlockRaw)
			heightBlockRaw.SetUid(dbClient.Messages[i].Uid)
			heightBlockRaw.Deserialize(dbClient.Messages[i].Message)
			heightBlockRaws = append(heightBlockRaws, heightBlockRaw)
		}
	}
	if len(heightBlockRaws) == 0 {
		return nil, nil
	}
	var newestHeightBlockRaw *HeightBlockRaw
	for _, heightBlockRaw := range heightBlockRaws {
		if newestHeightBlockRaw == nil || newestHeightBlockRaw.Height < heightBlockRaw.Height {
			newestHeightBlockRaw = heightBlockRaw
		}
	}
	if newestHeightBlockRaw == nil {
		return nil, nil
	}
	return newestHeightBlockRaw, nil
}

func GetHeightBlockRawSingle(height int64) (*HeightBlockRaw, error) {
	heightBlockRaws, err := GetHeightBlockRaw(height)
	if err != nil {
		return nil, jerr.Get("error getting height block raw", err)
	}
	if len(heightBlockRaws) == 0 {
		return nil, jerr.Get("error no height block raws found", client.EntryNotFoundError)
	} else if len(heightBlockRaws) > 1 {
		var hashStrings = make([]string, len(heightBlockRaws))
		for i := range heightBlockRaws {
			hashStrings[i] = hs.GetTxString(heightBlockRaws[i].BlockHash)
		}
		return nil, jerr.Getf(client.MultipleEntryError, "error more than 1 height block raw found: %d (%s)",
			len(heightBlockRaws), strings.Join(hashStrings, ", "))
	}
	return heightBlockRaws[0], nil
}

func GetHeightBlockRaw(height int64) ([]*HeightBlockRaw, error) {
	shardConfig := config.GetShardConfig(uint32(height), config.GetQueueShards())
	dbClient := client.NewClient(shardConfig.GetHost())
	err := dbClient.GetByPrefix(TopicHeightBlockRaw, jutil.GetInt64DataBig(height))
	if err != nil {
		return nil, jerr.Get("error getting height block raws for height from queue client", err)
	}
	var heightBlockRaws = make([]*HeightBlockRaw, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightBlockRaws[i] = new(HeightBlockRaw)
		heightBlockRaws[i].SetUid(dbClient.Messages[i].Uid)
		heightBlockRaws[i].Deserialize(dbClient.Messages[i].Message)
	}
	return heightBlockRaws, nil
}
