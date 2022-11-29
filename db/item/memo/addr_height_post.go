package memo

import (
	"context"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
)

type AddrHeightPost struct {
	Addr   [25]byte
	Height int64
	TxHash [32]byte
}

func (p *AddrHeightPost) GetTopic() string {
	return db.TopicMemoAddrHeightPost
}

func (p *AddrHeightPost) GetShard() uint {
	return client.GetByteShard(p.Addr[:])
}

func (p *AddrHeightPost) GetUid() []byte {
	return jutil.CombineBytes(
		p.Addr[:],
		jutil.ByteFlip(jutil.GetInt64DataBig(p.Height)),
		jutil.ByteReverse(p.TxHash[:]),
	)
}

func (p *AddrHeightPost) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(p.Addr[:], uid[:25])
	p.Height = jutil.GetInt64Big(jutil.ByteFlip(uid[25:33]))
	copy(p.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (p *AddrHeightPost) Serialize() []byte {
	return nil
}

func (p *AddrHeightPost) Deserialize([]byte) {}

func GetAddrHeightPosts(ctx context.Context, addrs [][25]byte) ([]*AddrHeightPost, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for _, addr := range addrs {
		shard := client.GetByteShard32(addr[:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addr[:])
	}
	shardConfigs := config.GetQueueShards()
	var addrPosts []*AddrHeightPost
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrHeightPost,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, jerr.Get("error getting db addr memo post by prefix", err)
		}
		for _, msg := range dbClient.Messages {
			var addrPost = new(AddrHeightPost)
			db.Set(addrPost, msg)
			addrPosts = append(addrPosts, addrPost)
		}
	}
	return addrPosts, nil
}
