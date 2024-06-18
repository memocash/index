package memo

import (
	"context"
	"fmt"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/client"
	"github.com/memocash/index/db/item/db"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/config"
	"time"
)

type AddrLinkRequested struct {
	Addr   [25]byte
	Seen   time.Time
	TxHash [32]byte
}

func (r *AddrLinkRequested) GetTopic() string {
	return db.TopicMemoAddrLinkRequested
}

func (r *AddrLinkRequested) GetShardSource() uint {
	return client.GenShardSource(r.Addr[:])
}

func (r *AddrLinkRequested) GetUid() []byte {
	return jutil.CombineBytes(
		r.Addr[:],
		jutil.GetTimeByteNanoBig(r.Seen),
		jutil.ByteReverse(r.TxHash[:]),
	)
}

func (r *AddrLinkRequested) SetUid(uid []byte) {
	if len(uid) != memo.AddressLength+memo.Int8Size+memo.TxHashLength {
		return
	}
	copy(r.Addr[:], uid[:25])
	r.Seen = jutil.GetByteTimeNanoBig(uid[25:33])
	copy(r.TxHash[:], jutil.ByteReverse(uid[33:65]))
}

func (r *AddrLinkRequested) Serialize() []byte {
	return nil
}

func (r *AddrLinkRequested) Deserialize([]byte) {}

func GetAddrLinkRequesteds(ctx context.Context, addrs [][25]byte) ([]*AddrLinkRequested, error) {
	var shardPrefixes = make(map[uint32][][]byte)
	for i := range addrs {
		shard := client.GenShardSource32(addrs[i][:])
		shardPrefixes[shard] = append(shardPrefixes[shard], addrs[i][:])
	}
	shardConfigs := config.GetQueueShards()
	var addrLinkRequesteds []*AddrLinkRequested
	for shard, prefixes := range shardPrefixes {
		shardConfig := config.GetShardConfig(shard, shardConfigs)
		dbClient := client.NewClient(shardConfig.GetHost())
		if err := dbClient.GetWOpts(client.Opts{
			Topic:    db.TopicMemoAddrLinkRequested,
			Prefixes: prefixes,
			Max:      client.ExLargeLimit,
			Context:  ctx,
		}); err != nil {
			return nil, fmt.Errorf("error getting db addr memo link requesteds by prefix; %w", err)
		}
		for _, msg := range dbClient.Messages {
			var addrLinkRequested = new(AddrLinkRequested)
			db.Set(addrLinkRequested, msg)
			addrLinkRequesteds = append(addrLinkRequesteds, addrLinkRequested)
		}
	}
	return addrLinkRequesteds, nil
}
