package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type AddrPost struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (p *AddrPost) GetTopic() string {
	return db.TopicMemoAddrPost
}

func (p *AddrPost) GetShard() uint {
	return client.GetByteShard(p.Addr[:])
}

func (p *AddrPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.GetTimeByteNanoBig(p.Seen),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrPost) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrPost) Serialize() []byte {
	return nil
}

func (p *AddrPost) Deserialize([]byte) {}

func GetAddrPosts(ctx context.Context, addrs [][25]byte, newest bool) ([]*AddrPost, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrPosts []*AddrPost
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrPost,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Newest:   newest,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db addr memo post by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var addrPost = new(AddrPost)
			db.Set(addrPost, msg)
			addrPosts = append(addrPosts, addrPost)
		}
	}
	return addrPosts, nil
}
