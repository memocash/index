package item

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/config"
	"strings"
	"time"
)

type HeightBlockShard struct {
	Height    int64
	BlockHash []byte
	Shard     uint
}

func (s HeightBlockShard) GetUid() []byte {
	return jutil.CombineBytes(
		jutil.GetUintData(s.Shard),
		jutil.GetInt64DataBig(s.Height),
		jutil.ByteReverse(s.BlockHash),
	)
}

func (s HeightBlockShard) GetShard() uint {
	return s.Shard
}

func (s HeightBlockShard) GetTopic() string {
	return TopicHeightBlockShard
}

func (s HeightBlockShard) Serialize() []byte {
	return nil
}

func (s *HeightBlockShard) SetUid(uid []byte) {
	if len(uid) != 44 {
		return
	}
	s.Shard = jutil.GetUint(uid[:4])
	s.Height = jutil.GetInt64Big(uid[4:12])
	s.BlockHash = jutil.ByteReverse(uid[12:44])
}

func (s *HeightBlockShard) Deserialize([]byte) {}

func GetRecentHeightBlockShard(shard uint) (*HeightBlockShard, error) {
	dbClient := client.NewClient(config.GetShardConfig(uint32(shard), config.GetQueueShards()).GetHost())
	if err := dbClient.Get(TopicHeightBlockShard, client.GetMaxStart(), false); err != nil {
		return nil, jerr.Getf(err, "error getting recent height block shard for shard: %d", shard)
	} else if len(dbClient.Messages) == 0 {
		return nil, nil
	} else if len(dbClient.Messages) > 1 {
		return nil, jerr.Newf("error unexpected number of recent height block shards returned: %d for shard: %d",
			len(dbClient.Messages), shard)
	}
	var heightBlockShard = new(HeightBlockShard)
	Set(heightBlockShard, dbClient.Messages[0])
	return heightBlockShard, nil
}

func GetHeightBlockShardSingle(shard uint, height int64) (*HeightBlockShard, error) {
	heightBlockShards, err := GetHeightBlockShard(shard, height)
	if err != nil {
		return nil, jerr.Get("error getting height block shard", err)
	} else if len(heightBlockShards) == 0 {
		return nil, jerr.Get("error no height block shards found", client.EntryNotFoundError)
	} else if len(heightBlockShards) > 1 {
		var hashStrings = make([]string, len(heightBlockShards))
		for i := range heightBlockShards {
			hashStrings[i] = hs.GetTxString(heightBlockShards[i].BlockHash)
		}
		return nil, jerr.Getf(client.MultipleEntryError, "error more than 1 height block shard found: %d (%s)",
			len(heightBlockShards), strings.Join(hashStrings, ", "))
	}
	return heightBlockShards[0], nil
}

func GetHeightBlockShard(shard uint, height int64) ([]*HeightBlockShard, error) {
	dbClient := client.NewClient(config.GetShardConfig(uint32(shard), config.GetQueueShards()).GetHost())
	prefix := jutil.CombineBytes(jutil.GetUintData(shard), jutil.GetInt64DataBig(height))
	if err := dbClient.GetByPrefix(TopicHeightBlockShard, prefix); err != nil {
		return nil, jerr.Get("error getting height block shards for height from queue client", err)
	}
	var heightBlockShards = make([]*HeightBlockShard, len(dbClient.Messages))
	for i := range dbClient.Messages {
		heightBlockShards[i] = new(HeightBlockShard)
		Set(heightBlockShards[i], dbClient.Messages[i])
	}
	return heightBlockShards, nil
}

func GetHeightBlockShardsAll(shard uint, startHeight int64, waitSingle bool) ([]*HeightBlockShard, error) {
	heightBlockShards, err := GetHeightBlockShardsAllLimit(shard, startHeight, waitSingle, client.LargeLimit, false)
	if err != nil {
		return nil, jerr.Get("error getting height block shards all large limit", err)
	}
	return heightBlockShards, nil
}

func GetHeightBlockShardsAllLimit(shard uint, startHeight int64, wait bool, limit uint32, newest bool) ([]*HeightBlockShard, error) {
	var timeout time.Duration
	if wait {
		timeout = time.Hour
	}
	var start []byte
	if startHeight != 0 {
		start = jutil.CombineBytes(jutil.GetUintData(shard), jutil.GetInt64DataBig(startHeight))
	}
	dbClient := client.NewClient(config.GetShardConfig(uint32(shard), config.GetQueueShards()).GetHost())
	if err := dbClient.GetWOpts(client.Opts{
		Topic:   TopicHeightBlockShard,
		Start:   start,
		Wait:    wait,
		Max:     limit,
		Newest:  newest,
		Timeout: timeout,
	}); err != nil {
		return nil, jerr.Get("error getting height block shards from queue client all", err)
	}
	var heightBlockShards []*HeightBlockShard
	for i := range dbClient.Messages {
		var heightBlockShard = new(HeightBlockShard)
		Set(heightBlockShard, dbClient.Messages[i])
		heightBlockShards = append(heightBlockShards, heightBlockShard)
	}
	return heightBlockShards, nil
}
